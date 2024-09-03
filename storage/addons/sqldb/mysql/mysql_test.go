package mysql

import (
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestBuildInsertQuery(t *testing.T) {
	ymlFile, err := ioutil.ReadFile("../testdata/sql.yaml")
	require.NoError(t, err)

	expected := []string{
		"INSERT INTO `table` (`field1`, `field2`) VALUES ('value1', 1)",
		"INSERT INTO `table` (`field1`, `field2`, `field3`) VALUES ('value2', 2, 2.5699477736545666)",
		"INSERT INTO `table` (`field1`, `field4`, `field5`) VALUES ('\"', false, NULL)",
		"INSERT INTO `table` (`field1`, `field5`) VALUES ('''', '[1,\"2\"]')",
	}

	ctx := loadContext{
		refsDefinition: make(map[string]row),
		refsInserted:   make(map[string]row),
	}

	require.NoError(t,
		loadYml("", ymlFile, &ctx),
	)

	for i, row := range ctx.tables[0].rows {
		query, err := buildInsertQuery(&ctx, "table", row)
		require.NoError(t, err)

		assert.Equal(t, expected[i], query)
	}
}

func TestLoadTablesShouldResolveRefs(t *testing.T) {
	yml, err := ioutil.ReadFile("../testdata/sql_refs.yaml")
	require.NoError(t, err)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() { _ = db.Close() }()

	ctx := loadContext{
		refsDefinition: make(map[string]row),
		refsInserted:   make(map[string]row),
	}

	if err = loadYml("", yml, &ctx); err != nil {
		t.Error(err)
		t.Fail()
	}

	mock.ExpectBegin()

	expectTruncate(mock, "table1")
	expectTruncate(mock, "table2")
	expectTruncate(mock, "table3")

	// table1
	expectInsert(t, mock,
		"table1",
		[]string{"f1", "f2"},
		"\\('value1', 'value2'\\)",
		[]string{"value1", "value2"},
	)

	// table2
	expectInsert(t, mock,
		"table2",
		[]string{"f1", "f2"},
		"\\('value2', 'value1'\\)",
		[]string{"value2", "value1"},
	)

	// table1
	expectInsert(t, mock,
		"table3",
		[]string{"f1", "f2"},
		"\\('value1', 'value2'\\)",
		[]string{"value1", "value2"},
	)

	mock.ExpectCommit()

	err = loadTables(&ctx, db)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		t.Fail()
	}
}

func TestLoadTablesShouldExtendRows(t *testing.T) {
	yml, err := ioutil.ReadFile("../testdata/sql_extend.yaml")
	require.NoError(t, err)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func() { _ = db.Close() }()

	ctx := loadContext{
		refsDefinition: make(map[string]row),
		refsInserted:   make(map[string]row),
	}

	if err = loadYml("", yml, &ctx); err != nil {
		t.Error(err)
		t.Fail()
	}

	mock.ExpectBegin()

	expectTruncate(mock, "table1")
	expectTruncate(mock, "table2")
	expectTruncate(mock, "table3")

	// table1
	expectInsert(t, mock,
		"table1",
		[]string{"f1", "f2"},
		"\\('value1', 'value2'\\)",
		[]string{"value1", "value2"},
	)

	// table2
	expectInsert(t, mock,
		"table2",
		[]string{"f1", "f2", "f3"},
		"\\('value1 overwritten', 'value2', "+`\("1" \|\| "2" \|\| 3 \+ 5\)\)$`,
		[]string{"value1 overwritten", "value2", `1`},
	)

	// table3, data 1
	expectInsert(t, mock,
		"table3",
		[]string{"f1", "f2", "f3"},
		"\\('value1 overwritten', 'value2', "+`\("1" \|\| "2" \|\| 3 \+ 5\)\)$`,
		[]string{"value1 overwritten", "value2", `1`},
	)

	// table3, data 2
	expectInsert(t, mock,
		"table3",
		[]string{"f1", "f2"},
		"\\('tplVal1', 'tplVal2'\\)",
		[]string{"tplVal1", "tplVal2"},
	)

	mock.ExpectCommit()

	err = loadTables(&ctx, db)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		t.Fail()
	}
}

var idCounter int64

func expectInsert(
	t *testing.T,
	mock sqlmock.Sqlmock,
	table string,
	fields []string,
	valuesToInsert string,
	valuesResult []string,
) {
	t.Helper()

	idCounter++

	mock.ExpectExec(
		fmt.Sprintf("^INSERT INTO `%s` %s VALUES %s$",
			table,
			fieldsToDbStr(fields),
			valuesToInsert,
		),
	).
		WillReturnResult(sqlmock.NewResult(idCounter, 1))

	var valuesRow []driver.Value
	for _, v := range valuesResult {
		valuesRow = append(valuesRow, driver.Value(v))
	}

	mock.ExpectQuery(fmt.Sprintf("^SELECT \\* FROM `%s` WHERE `id` = \\?$", table)).
		WithArgs(idCounter).
		WillReturnRows(
			sqlmock.NewRows(fields).
				AddRow(valuesRow...),
		)
}

func expectTruncate(mock sqlmock.Sqlmock, table string) {
	mock.ExpectExec(fmt.Sprintf("^TRUNCATE TABLE `%s`$", table)).
		WillReturnResult(sqlmock.NewResult(0, 0))
}

func fieldsToDbStr(values []string) string {
	quotedVals := make([]string, len(values))

	for i, val := range values {
		quotedVals[i] = "`" + val + "`"
	}

	return "\\(" + strings.Join(quotedVals, ", ") + "\\)"
}
