package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type row map[string]interface{}

type table []row

type rowsDict map[string]row

type fixture struct {
	Inherits  []string
	Tables    yaml.MapSlice
	Templates yaml.MapSlice
}

type loadedTable struct {
	name tableName
	rows table
}
type tableName struct {
	name   string
	schema string
}

func newTableName(source string) tableName {
	parts := strings.SplitN(source, ".", 2)
	switch {
	case len(parts) == 1:
		parts = append(parts, parts[0])
		fallthrough
	case parts[0] == "":
		parts[0] = "public"
	}
	lt := tableName{schema: parts[0], name: parts[1]}

	return lt
}

func (t *tableName) getFullName() string {
	return fmt.Sprintf("%q.%q", t.schema, t.name)
}

type loadContext struct {
	files          []string
	tables         []loadedTable
	refsDefinition rowsDict
	refsInserted   rowsDict
}

func LoadFixtures(db *sql.DB, location string, names []string) error {
	ctx := loadContext{
		refsDefinition: make(rowsDict),
		refsInserted:   make(rowsDict),
	}
	// gather data from files
	for _, name := range names {
		err := loadFile(location, name, &ctx)
		if err != nil {
			return fmt.Errorf("unable to load fixture %s: %s", name, err.Error())
		}
	}

	return loadTables(&ctx, db)
}

func ExecuteQuery(db *sql.DB, query string) ([]json.RawMessage, error) {
	dbResponse := []json.RawMessage{}

	if idx := strings.IndexByte(query, ';'); idx >= 0 {
		query = query[:idx]
	}

	rows, err := db.Query(fmt.Sprintf("SELECT row_to_json(rows) FROM (%s) rows;", query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var jsonString string
		err = rows.Scan(&jsonString)
		if err != nil {
			return nil, err
		}
		dbResponse = append(dbResponse, json.RawMessage(jsonString))
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return dbResponse, nil
}

func loadFile(location, name string, ctx *loadContext) error {
	candidates := []string{
		location + "/" + name,
		location + "/" + name + ".yml",
		location + "/" + name + ".yaml",
	}
	var err error
	var file string
	for _, candidate := range candidates {
		if _, err = os.Stat(candidate); err == nil {
			file = candidate

			break
		}
	}
	if err != nil {
		return err
	}
	// skip previously loaded files
	if inArray(file, ctx.files) {
		return nil
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	ctx.files = append(ctx.files, file)

	return loadYml(location, data, ctx)
}

func loadYml(location string, data []byte, ctx *loadContext) error {
	// read yml into struct
	var loadedFixture fixture
	if err := yaml.Unmarshal(data, &loadedFixture); err != nil {
		return err
	}

	// load inherits
	for _, inheritFile := range loadedFixture.Inherits {
		if err := loadFile(location, inheritFile, ctx); err != nil {
			return err
		}
	}

	// loadedFixture.templates
	// yaml.MapSlice{
	//    string => yaml.MapSlice{
	//        string => interface{}
	//    }
	// }
	for _, template := range loadedFixture.Templates {
		name := template.Key.(string)
		if _, ok := ctx.refsDefinition[name]; ok {
			return fmt.Errorf("unable to load template %s: duplicating ref name", name)
		}
		fields := template.Value.(yaml.MapSlice)
		row := make(row, len(fields))
		for _, field := range fields {
			key := field.Key.(string)
			row[key] = field.Value
		}
		if base, ok := row["$extend"]; ok {
			base := base.(string)
			baseRow, err := resolveReference(ctx.refsDefinition, base)
			if err != nil {
				return err
			}
			for k, v := range row {
				baseRow[k] = v
			}
			row = baseRow
		}
		ctx.refsDefinition[name] = row
	}

	// loadedFixture.tables
	// yaml.MapSlice{
	//    string => []interface{
	//        yaml.MapSlice{
	//            string => interface{}
	//        }
	//    }
	// }
	for _, sourceTable := range loadedFixture.Tables {
		sourceRows, ok := sourceTable.Value.([]interface{})
		if !ok {
			return errors.New("expected array at root level")
		}
		rows := make(table, len(sourceRows))
		for i := range sourceRows {
			sourceFields := sourceRows[i].(yaml.MapSlice)
			fields := make(row, len(sourceFields))
			for j := range sourceFields {
				fields[sourceFields[j].Key.(string)] = sourceFields[j].Value
			}
			rows[i] = fields
		}
		lt := loadedTable{
			name: newTableName(sourceTable.Key.(string)),
			rows: rows,
		}
		ctx.tables = append(ctx.tables, lt)
	}

	return nil
}

func loadTables(ctx *loadContext, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// truncate first
	if err := truncateTables(tx, ctx.tables...); err != nil {
		return err
	}

	// then load data
	for _, lt := range ctx.tables {
		if len(lt.rows) == 0 {
			continue
		}
		if err := loadTable(ctx, tx, lt.name, lt.rows); err != nil {
			return fmt.Errorf("failed to load table '%s' because:\n%s", lt.name.getFullName(), err)
		}
	}
	// alter the sequences so they contain max id + 1
	if err := fixSequences(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// truncateTables truncates table
func truncateTables(tx *sql.Tx, tables ...loadedTable) error {
	set := make(map[string]struct{})
	tablesToTruncate := make([]string, 0, len(tables))
	for _, t := range tables {
		tableName := t.name.getFullName()
		if _, ok := set[tableName]; ok {
			// already truncated
			continue
		}

		tablesToTruncate = append(tablesToTruncate, tableName)
		set[tableName] = struct{}{}
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(tablesToTruncate, ","))
	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func loadTable(ctx *loadContext, tx *sql.Tx, t tableName, rows table) error {
	// $extend keyword allows to import values from a named row
	for i, row := range rows {
		if _, ok := row["$extend"]; !ok {
			continue
		}
		base := row["$extend"].(string)
		baseRow, err := resolveReference(ctx.refsDefinition, base)
		if err != nil {
			return err
		}
		for k, v := range row {
			baseRow[k] = v
		}
		rows[i] = baseRow
	}
	// build SQL
	query, err := buildInsertQuery(ctx, t, rows)
	if err != nil {
		return err
	}
	// issuing query
	insertedRows, err := tx.Query(query)
	if err != nil {
		return err
	}
	defer func() { _ = insertedRows.Close() }()
	// reading results
	// here I assume that returning rows go in the same
	// order as values were passed to INSERT statement
	for _, row := range rows {
		if !insertedRows.Next() {
			break
		}
		if name, ok := row["$name"]; ok {
			name := name.(string)
			if _, ok := ctx.refsDefinition[name]; ok {
				return fmt.Errorf("duplicating ref name %s", name)
			}
			// read values
			var rowJSON string
			if err := insertedRows.Scan(&rowJSON); err != nil {
				return err
			}
			// decode json
			values := make(map[string]interface{})
			if err := json.Unmarshal([]byte(rowJSON), &values); err != nil {
				return err
			}
			// add to references
			ctx.refsDefinition[name] = row
			ctx.refsInserted[name] = values
		}
	}

	// iterate through any remaining rows and check for an error
	for insertedRows.Next() {
		continue
	}
	if err := insertedRows.Err(); err != nil {
		return fmt.Errorf("failed to execute query. DB returned error:\n%s", err)
	}

	return err
}

func fixSequences(tx *sql.Tx) error {
	query := `
DO $$
DECLARE
    r record;
BEGIN
    FOR r IN (
        SELECT 'SELECT SETVAL(' || quote_literal(quote_ident(seq_ns.nspname) || '.' || quote_ident(seq.relname))
            || ', COALESCE(MAX(' || quote_ident(col.attname) || '), 1) ) FROM '
            || quote_ident(tbl_ns.nspname) || '.' || quote_ident(tbl.relname) AS q
        FROM pg_class seq
            JOIN pg_namespace seq_ns ON (seq.relnamespace = seq_ns.oid)
            JOIN pg_depend dep ON (dep.objid = seq.oid)
            JOIN pg_class tbl ON (dep.refobjid = tbl.oid)
            JOIN pg_namespace tbl_ns ON (tbl.relnamespace = tbl_ns.oid)
            JOIN pg_attribute col ON (col.attrelid = tbl.oid AND dep.refobjsubid = col.attnum)
        WHERE
            seq.relkind = 'S'
        ORDER BY seq.relname
    ) LOOP
        EXECUTE r.q;
    END LOOP;
END$$
`
	_, err := tx.Exec(query)

	return err
}

// buildInsertQuery builds SQL query for data insertion
// based on values read from yaml
func buildInsertQuery(ctx *loadContext, t tableName, rows table) (string, error) {
	// first pass, collecting all the fields
	var fields []string
	fieldPresence := make(map[string]bool)
	for _, row := range rows {
		for name := range row {
			if name != "" && name[0] == '$' {
				continue
			}
			if _, ok := fieldPresence[name]; !ok {
				fieldPresence[name] = true
				fields = append(fields, name)
			}
		}
	}
	sort.Strings(fields)
	// second pass, collecting values
	dbValues := make([]string, len(rows))
	for i, row := range rows {
		dbValuesRow := make([]string, len(fields))
		for k, name := range fields {
			value, present := row[name]
			if !present {
				dbValuesRow[k] = "default" // default is a PostgreSQL keyword

				continue
			}
			// resolve references
			if stringValue, ok := value.(string); ok {
				if stringValue != "" && stringValue[0] == '$' {
					var err error
					dbValuesRow[k], err = resolveExpression(stringValue, ctx)
					if err != nil {
						return "", err
					}

					continue
				}
			}
			dbValue, err := toDbValue(value)
			if err != nil {
				return "", fmt.Errorf("unable to process %s value (row %d of %s): %s", name, i, t.getFullName(), err.Error())
			}
			dbValuesRow[k] = dbValue
		}
		dbValues[i] = "(" + strings.Join(dbValuesRow, ", ") + ")"
	}
	// quote fields
	for i, field := range fields {
		fields[i] = "\"" + field + "\""
	}

	query := "INSERT INTO %s AS row (%s) VALUES %s RETURNING row_to_json(row)"

	return fmt.Sprintf(query, t.getFullName(), strings.Join(fields, ", "), strings.Join(dbValues, ", ")), nil
}

// resolveExpression converts expressions starting with dollar sign into a value
// supporting expressions:
// - $eval()               - executes an SQL expression, e.g. $eval(CURRENT_DATE)
// - $recordName.fieldName - using value of previously inserted named record
func resolveExpression(expr string, ctx *loadContext) (string, error) {
	if expr[:5] == "$eval" {
		re := regexp.MustCompile(`^\$eval\((.+)\)$`)
		if matches := re.FindStringSubmatch(expr); matches != nil {
			return "(" + matches[1] + ")", nil
		}

		return "", fmt.Errorf("icorrect $eval() usage: %s", expr)
	}

	value, err := resolveFieldReference(ctx.refsInserted, expr)
	if err != nil {
		return "", nil
	}

	return toDbValue(value)
}

// resolveReference finds previously stored reference by its name
func resolveReference(refs rowsDict, refName string) (row, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference %s", refName)
	}
	// make a copy of referencing data to prevent spoiling the source
	// by the way removing $-records from base row
	targetCopy := make(row, len(target))
	for k, v := range target {
		if k == "" || k[0] != '$' {
			targetCopy[k] = v
		}
	}

	return targetCopy, nil
}

// resolveFieldReference finds previously stored reference by name
// and return value of its field
func resolveFieldReference(refs rowsDict, ref string) (interface{}, error) {
	parts := strings.SplitN(ref, ".", 2)
	if len(parts) < 2 || len(parts[0]) < 2 || len(parts[1]) < 1 {
		return nil, fmt.Errorf("invalid reference %s, correct form is $refName.field", ref)
	}
	// remove leading $
	refName := parts[0][1:]
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference %s", refName)
	}
	value, ok := target[parts[1]]
	if !ok {
		return nil, fmt.Errorf("undefined reference field %s", parts[1])
	}

	return value, nil
}

// inArray checks whether the needle is present in haystack slice
func inArray(needle string, haystack []string) bool {
	for _, e := range haystack {
		if needle == e {
			return true
		}
	}

	return false
}

// toDbValue prepares value to be passed in SQL query
// with respect to its type and converts it to string
func toDbValue(value interface{}) (string, error) {
	if value == nil {
		return "NULL", nil
	}
	if value, ok := value.(string); ok {
		return quoteLiteral(value), nil
	}
	if value, ok := value.(int); ok {
		return strconv.Itoa(value), nil
	}
	if value, ok := value.(float64); ok {
		return strconv.FormatFloat(value, 'g', -1, 64), nil
	}
	if value, ok := value.(bool); ok {
		return strconv.FormatBool(value), nil
	}
	// the value is either slice or map, so insert it as JSON string
	// fixme: marshaller doesn't know how to encode map[interface{}]interface{}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return quoteLiteral(string(encoded)), nil
}

// quoteLiteral properly escapes string to be safely
// passed as a value in SQL query
func quoteLiteral(s string) string {
	var p string
	if strings.Contains(s, `\`) {
		p = "E"
	}
	s = strings.ReplaceAll(s, `'`, `''`)
	s = strings.ReplaceAll(s, `\`, `\\`)

	return p + `'` + s + `'`
}
