package fixtures

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

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

type row map[string]interface{}

type table []row

type rowsDict map[string]row

type fixture struct {
	Version   string
	Inherits  []string
	Tables    yaml.MapSlice
	Templates yaml.MapSlice
}

type loadedTable struct {
	Name string
	Rows table
}

type loadContext struct {
	files          []string
	tables         []loadedTable
	refsDefinition rowsDict
	refsInserted   rowsDict
}

type Config struct {
	DB       *sql.DB
	Location string
	Debug    bool
}

type Loader struct {
	db       *sql.DB
	location string
	debug    bool
}

func NewLoader(config *Config) *Loader {
	return &Loader{
		db:       config.DB,
		location: strings.TrimRight(config.Location, "/"),
		debug:    config.Debug,
	}
}

func (f *Loader) Load(names []string) error {
	ctx := loadContext{
		refsDefinition: make(rowsDict),
		refsInserted:   make(rowsDict),
	}
	// gather data from files
	for _, name := range names {
		err := f.loadFile(name, &ctx)
		if err != nil {
			return fmt.Errorf("unable to load fixture %s: %s", name, err.Error())
		}
	}
	return f.loadTables(&ctx)
}

func (f *Loader) loadFile(name string, ctx *loadContext) error {
	candidates := []string{
		f.location + "/" + name,
		f.location + "/" + name + ".yml",
		f.location + "/" + name + ".yaml",
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
	if inArray(file, &(*ctx).files) {
		return nil
	}
	if f.debug {
		fmt.Println("Loading", file)
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	(*ctx).files = append((*ctx).files, file)
	return f.loadYml(data, ctx)
}

func (f *Loader) loadYml(data []byte, ctx *loadContext) error {
	// read yml into struct
	var loadedFixture fixture
	if err := yaml.Unmarshal(data, &loadedFixture); err != nil {
		return err
	}

	// load inherits
	for _, inheritFile := range loadedFixture.Inherits {
		if err := f.loadFile(inheritFile, ctx); err != nil {
			return err
		}
	}

	// loadedFixture.Templates
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
			value, _ := field.Value.(interface{})
			row[key] = value
		}
		if base, ok := row["$extend"]; ok {
			base := base.(string)
			baseRow, err := f.resolveReference(ctx.refsDefinition, base)
			if err != nil {
				return err
			}
			for k, v := range row {
				baseRow[k] = v
			}
			row = baseRow
		}
		ctx.refsDefinition[name] = row
		if f.debug {
			rowJson, _ := json.Marshal(row)
			fmt.Printf("Populating ref %s as %s from template\n", name, string(rowJson))
		}
	}

	// loadedFixture.Tables
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
			Name: sourceTable.Key.(string),
			Rows: rows,
		}
		(*ctx).tables = append((*ctx).tables, lt)
	}
	return nil
}

func (f *Loader) loadTables(ctx *loadContext) error {
	tx, err := f.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// truncate first
	truncatedTables := make(map[string]bool)
	for _, lt := range ctx.tables {
		if _, ok := truncatedTables[lt.Name]; ok {
			// already truncated
			continue
		}
		if err := f.truncateTable(lt.Name); err != nil {
			return err
		}
		truncatedTables[lt.Name] = true
	}
	// then load data
	for _, lt := range ctx.tables {
		if len(lt.Rows) == 0 {
			continue
		}
		if err := f.loadTable(ctx, lt.Name, lt.Rows); err != nil {
			return err
		}
	}
	// alter the sequences so they contain max id + 1
	if err := f.fixSequences(); err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// truncateTable truncates table
func (f *Loader) truncateTable(name string) error {
	query := fmt.Sprintf("TRUNCATE TABLE \"%s\" CASCADE", name)
	if f.debug {
		fmt.Println("Issuing SQL:", query)
	}
	_, err := f.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (f *Loader) loadTable(ctx *loadContext, t string, rows table) error {
	// $extend keyword allows to import values from a named row
	for i, row := range rows {
		if base, ok := row["$extend"]; ok {
			base := base.(string)
			baseRow, err := f.resolveReference(ctx.refsDefinition, base)
			if err != nil {
				return err
			}
			for k, v := range row {
				baseRow[k] = v
			}
			rows[i] = baseRow
		}
	}
	// build SQL
	query, err := f.buildInsertQuery(ctx, t, rows)
	if err != nil {
		return err
	}
	if f.debug {
		fmt.Println("Issuing SQL:", query)
	}
	// issuing query
	insertedRows, err := f.db.Query(query)
	if err != nil {
		return err
	}
	defer insertedRows.Close()
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
			var rowJson string
			if err := insertedRows.Scan(&rowJson); err != nil {
				return err
			}
			// decode json
			values := make(map[string]interface{})
			if err := json.Unmarshal([]byte(rowJson), &values); err != nil {
				return err
			}
			// add to references
			ctx.refsDefinition[name] = row
			if f.debug {
				rowJson, _ := json.Marshal(values)
				fmt.Printf("Populating ref %s as %s from row definition\n", name, string(rowJson))
			}
			ctx.refsInserted[name] = values
			if f.debug {
				rowJson, _ := json.Marshal(values)
				fmt.Printf("Populating ref %s as %s from inserted values\n", name, string(rowJson))
			}
		}
	}
	return err
}

func (f *Loader) fixSequences() error {
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
	if f.debug {
		fmt.Println("Issuing SQL:", query)
	}
	_, err := f.db.Exec(query)
	return err
}

// buildInsertQuery builds SQL query for data insertion
// based on values read from yaml
func (f *Loader) buildInsertQuery(ctx *loadContext, t string, rows table) (string, error) {
	// first pass, collecting all the fields
	var fields []string
	fieldPresence := make(map[string]bool)
	for _, row := range rows {
		for name, _ := range row {
			if len(name) > 0 && name[0] == '$' {
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
			value, _ := row[name]
			// resolve references
			if stringValue, ok := value.(string); ok {
				if len(stringValue) > 0 && stringValue[0] == '$' {
					var err error
					dbValuesRow[k], err = f.resolveExpression(stringValue, ctx)
					if err != nil {
						return "", err
					}
					continue
				}
			}
			dbValue, err := toDbValue(value)
			if err != nil {
				return "", fmt.Errorf("unable to process %s value (row %d of %s): %s", name, i, t, err.Error())
			}
			dbValuesRow[k] = dbValue
		}
		dbValues[i] = "(" + strings.Join(dbValuesRow, ", ") + ")"
	}
	// quote fields
	for i, field := range fields {
		fields[i] = "\"" + field + "\""
	}
	query := "INSERT INTO \"%s\" (%s) VALUES %s RETURNING row_to_json(\"%[1]s\")"
	return fmt.Sprintf(query, t, strings.Join(fields, ", "), strings.Join(dbValues, ", ")), nil
}

// resolveExpression converts expressions starting with dollar sign into a value
// supporting expressions:
// - $eval()               - executes an SQL expression, e.g. $eval(CURRENT_DATE)
// - $recordName.fieldName - using value of previously inserted named record
func (f *Loader) resolveExpression(expr string, ctx *loadContext) (string, error) {
	if expr[:5] == "$eval" {
		re := regexp.MustCompile(`^\$eval\((.+)\)$`)
		if matches := re.FindStringSubmatch(expr); matches != nil {
			return "(" + matches[1] + ")", nil
		} else {
			return "", fmt.Errorf("icorrect $eval() usage: %s", expr)
		}
	} else {
		value, err := f.resolveFieldReference(ctx.refsInserted, expr)
		if err != nil {
			return "", nil
		}
		return toDbValue(value)
	}
}

// resolveReference finds previously stored reference by its name
func (f *Loader) resolveReference(refs rowsDict, refName string) (row, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference %s", refName)
	}
	// make a copy of referencing data to prevent spoiling the source
	// by the way removing $-records from base row
	targetCopy := make(row, len(target))
	for k, v := range target {
		if len(k) == 0 || k[0] != '$' {
			targetCopy[k] = v
		}
	}
	return targetCopy, nil
}

// resolveFieldReference finds previously stored reference by name
// and return value of its field
func (f *Loader) resolveFieldReference(refs rowsDict, ref string) (interface{}, error) {
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
func inArray(needle string, haystack *[]string) bool {
	for _, e := range *haystack {
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
	s = strings.Replace(s, `'`, `''`, -1)
	s = strings.Replace(s, `\`, `\\`, -1)
	return p + `'` + s + `'`
}
