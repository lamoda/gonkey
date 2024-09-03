package mysql

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

const errNoIDColumn = "Error 1054: Unknown column 'id' in 'where clause'"

type row map[string]interface{}

type table []row

type rowsDict map[string]row

type fixture struct {
	Inherits  []string
	Tables    yaml.MapSlice
	Templates yaml.MapSlice
}

type loadedTable struct {
	name string
	rows table
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
	return nil, fmt.Errorf("not implemented")
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
			name: sourceTable.Key.(string),
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
	truncatedTables := make(map[string]bool)
	for _, lt := range ctx.tables {
		if _, ok := truncatedTables[lt.name]; ok {
			// already truncated
			continue
		}
		if err := truncateTable(tx, lt.name); err != nil {
			return err
		}
		truncatedTables[lt.name] = true
	}

	// then load data
	for _, lt := range ctx.tables {
		if len(lt.rows) == 0 {
			continue
		}
		if err := loadTable(tx, ctx, lt.name, lt.rows); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func truncateTable(tx *sql.Tx, name string) error {
	query := fmt.Sprintf("TRUNCATE TABLE `%s`", name)

	_, err := tx.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func loadTable(tx *sql.Tx, ctx *loadContext, t string, rows table) error {
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

	// issuing query
	for _, row := range rows {
		if err := loadRow(tx, ctx, t, row); err != nil {
			return err
		}
	}

	return nil
}

func loadRow(tx *sql.Tx, ctx *loadContext, t string, row row) error {
	query, err := buildInsertQuery(ctx, t, row)
	if err != nil {
		return err
	}

	insertRes, err := tx.Exec(query)
	if err != nil {
		return err
	}

	// find inserted rows
	insertedRow, err := insertedRows(tx, insertRes, t)
	defer func() {
		if insertedRow != nil {
			_ = insertedRow.Close()
		}
	}()

	if err != nil {
		return err
	}

	// TODO: we couldn't get insertedRow because don't know Primary Key
	if insertedRow == nil {
		return nil
	}

	if !insertedRow.Next() {
		return errors.New("can't get inserted row")
	}

	if name, ok := row["$name"]; ok {
		name := name.(string)
		if _, ok := ctx.refsDefinition[name]; ok {
			return fmt.Errorf("duplicating ref name %s", name)
		}

		insertedRowValue, err := fetchRow(insertedRow)
		if err != nil {
			return err
		}

		// add to references
		ctx.refsDefinition[name] = row
		ctx.refsInserted[name] = insertedRowValue
	}

	return nil
}

func fetchRow(rows *sql.Rows) (row, error) {
	res := make(row)

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	rawResult := make([]sql.RawBytes, len(cols))

	dest := make([]interface{}, len(cols))
	for i := range rawResult {
		dest[i] = &rawResult[i]
	}

	// read values
	if err := rows.Scan(dest...); err != nil {
		return nil, err
	}

	for i, raw := range rawResult {
		if raw == nil {
			res[cols[i]] = "NULL"
		} else {
			res[cols[i]] = string(raw)
		}
	}

	return res, nil
}

func insertedRows(tx *sql.Tx, insertRes sql.Result, t string) (*sql.Rows, error) {
	lastID, err := insertRes.LastInsertId()
	if err != nil {
		return nil, err
	}

	//nolint:gosec // Obviously shouldn't be used with production DB.
	query := fmt.Sprintf("SELECT * FROM `%s` WHERE `id` = ?", t)

	rows, err := tx.Query(query, lastID)
	if err != nil {
		// TODO: now we can take inserted rows only if they have column 'id'
		//  later we can add possibility to specify name of PK column in fixture definition
		//  Also, it's weak error check
		if err.Error() == errNoIDColumn {
			return nil, nil
		}

		return nil, err
	}

	return rows, nil
}

// buildInsertQuery builds SQL query for data insertion
// based on values read from yaml
func buildInsertQuery(ctx *loadContext, t string, row row) (string, error) {
	fields := make([]string, 0, len(row))

	for name := range row {
		if strings.HasPrefix(name, "$") {
			continue
		}
		fields = append(fields, name)
	}

	sort.Strings(fields)

	values := make([]string, len(fields))

	for i, name := range fields {
		val := row[name]

		v, err := rowInsertValue(ctx, val)
		if err != nil {
			return "", fmt.Errorf(
				"unable to process %s value (of %s): %s",
				name, t, err.Error(),
			)
		}

		values[i] = v
	}

	// quote fields
	for i, field := range fields {
		fields[i] = "`" + field + "`"
	}

	query := "INSERT INTO `%s` (%s) VALUES %s"

	return fmt.Sprintf(
		query,
		t,
		strings.Join(fields, ", "),
		"("+strings.Join(values, ", ")+")",
	), nil
}

func rowInsertValue(ctx *loadContext, val interface{}) (string, error) {
	// resolve references
	if stringValue, ok := val.(string); ok {
		if strings.HasPrefix(stringValue, "$") {
			v, err := resolveExpression(stringValue, ctx)
			if err != nil {
				return "", err
			}

			return v, nil
		}
	}

	dbValue, err := toDbValue(val)
	if err != nil {
		return "", err
	}

	return dbValue, nil
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
		return "", err
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
	s = strings.ReplaceAll(s, `'`, `''`)
	s = strings.ReplaceAll(s, `\`, `\\`)
	return "'" + s + "'"
}
