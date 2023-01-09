package postgres

import (
	"database/sql"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestBuildInsertQuery(t *testing.T) {
	yml, err := ioutil.ReadFile("../testdata/sql.yaml")
	require.NoError(t, err)

	expected := "INSERT INTO \"public\".\"table\" AS row (\"field1\", \"field2\", \"field3\", \"field4\", \"field5\") VALUES " +
		"('value1', 1, default, default, default), " +
		"('value2', 2, 2.5699477736545666, default, default), " +
		"('\"', default, default, false, NULL), " +
		"('''', default, default, default, '[1,\"2\"]') " +
		"RETURNING row_to_json(row)"

	ctx := loadContext{
		refsDefinition: make(map[string]row),
		refsInserted:   make(map[string]row),
	}

	l := New(&sql.DB{}, "", false)
	err = l.loadYml(yml, &ctx)
	require.NoError(t, err)

	query, err := l.buildInsertQuery(&ctx, newTableName("table"), ctx.tables[0].rows)

	if err != nil {
		t.Error("must not produce error, error:", err.Error())
		t.Fail()
	}

	if query != expected {
		t.Error("must generate proper SQL, got result:", query)
		t.Fail()
	}
}

func TestLoadTablesShouldResolveSchema(t *testing.T) {
	yml, err := ioutil.ReadFile("../testdata/sql_schema.yaml")
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

	l := New(db, "", true)

	err = l.loadYml([]byte(yml), &ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	mock.ExpectBegin()

	mock.ExpectExec("^TRUNCATE TABLE \"schema1\".\"table1\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"schema2\".\"table2\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table3\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	q := `^INSERT INTO "schema1"."table1" AS row \("f1", "f2"\) VALUES ` +
		`\('value1', 'value2'\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1\",\"f2\":\"value2\"}"),
		)

	q = `^INSERT INTO "schema2"."table2" AS row \("f1", "f2"\) VALUES ` +
		`\('value3', 'value4'\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value3\",\"f2\":\"value4\"}"),
		)

	q = `^INSERT INTO "public"."table3" AS row \("f1", "f2"\) VALUES ` +
		`\('value5', 'value6'\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value5\",\"f2\":\"value6\"}"),
		)

	mock.ExpectExec("^DO").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectCommit()

	err = l.loadTables(&ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		t.Fail()
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

	l := New(db, "", true)

	err = l.loadYml([]byte(yml), &ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	mock.ExpectBegin()

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table1\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table2\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table3\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	q := `^INSERT INTO "public"."table1" AS row \("f1", "f2"\) VALUES ` +
		`\('value1', 'value2'\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1\",\"f2\":\"value2\"}"),
		)

	q = `^INSERT INTO "public"."table2" AS row \("f1", "f2"\) VALUES ` +
		`\('value2', 'value1'\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value2\",\"f2\":\"value1\"}"),
		)

	q = `^INSERT INTO "public"."table3" AS row \("f1", "f2"\) VALUES ` +
		`\('value1', 'value2'\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1\",\"f2\":\"value2\"}"),
		)

	mock.ExpectExec("^DO").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectCommit()

	err = l.loadTables(&ctx)
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

	l := New(db, "", true)

	err = l.loadYml([]byte(yml), &ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	mock.ExpectBegin()

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table1\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table2\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table3\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	q := `^INSERT INTO "public"."table1" AS row \("f1", "f2"\) VALUES ` +
		`\('value1', 'value2'\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1\",\"f2\":\"value2\"}"),
		)

	q = `^INSERT INTO "public"."table2" AS row \("f1", "f2", "f3"\) VALUES ` +
		`\('value1 overwritten', 'value2', \(\"1\" \|\| \"2\" \|\| 3 \+ 5\)\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1 overwritten\",\"f2\":\"value2\",\"f3\":\"value3\"}"),
		)

	q = `^INSERT INTO "public"."table3" AS row \("f1", "f2", "f3"\) VALUES ` +
		`\('value1 overwritten', 'value2', \(\"1\" \|\| \"2\" \|\| 3 \+ 5\)\), ` +
		`\('tplVal1', 'tplVal2', default\) ` +
		`RETURNING row_to_json\(row\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1 overwritten\",\"f2\":\"value2\",\"f3\":\"value3\"}").
				AddRow("{\"f1\":\"tplValue1\",\"f2\":\"tplValue2\",\"f3\":null}"),
		)

	mock.ExpectExec("^DO").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectCommit()

	err = l.loadTables(&ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		t.Fail()
	}
}
