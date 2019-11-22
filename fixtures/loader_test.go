package fixtures

import (
	"testing"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestBuildInsertQuery(t *testing.T) {
	yml := `
tables:
  table:
    - field1: "value1"
      field2: 1
    - field1: "value2"
      field2: 2
      field3: 2.569947773654566473
    - field4: false
      field5: null
      field1: '"'
    - field1: "'"
      field5:
        - 1
        - '2'
`
	expected := "INSERT INTO \"table\" AS table_table_gonkey (\"field1\", \"field2\", \"field3\", \"field4\", \"field5\") VALUES " +
		"('value1', 1, default, default, default), " +
		"('value2', 2, 2.5699477736545666, default, default), " +
		"('\"', default, default, false, NULL), " +
		"('''', default, default, default, '[1,\"2\"]') " +
		"RETURNING row_to_json(table_table_gonkey)"

	ctx := loadContext{
		refsDefinition: make(map[string]row),
		refsInserted:   make(map[string]row),
	}

	l := NewLoader(&Config{})
	l.loadYml([]byte(yml), &ctx)

	query, err := l.buildInsertQuery(&ctx, "table", ctx.tables[0].Rows)

	if err != nil {
		t.Error("must not produce error, error:", err.Error())
		t.Fail()
	}

	if query != expected {
		t.Error("must generate proper SQL, got result:", query)
		t.Fail()
	}
}

func TestLoadTablesShouldResolveRefs(t *testing.T) {
	yml := `
tables:
  table1:
    - $name: ref1
      f1: value1
      f2: value2

  table2:
    - $name: ref2
      f1: $ref1.f2
      f2: $ref1.f1

  table3:
    - f1: $ref1.f1
      f2: $ref2.f1
`

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctx := loadContext{
		refsDefinition: make(map[string]row),
		refsInserted:   make(map[string]row),
	}

	l := NewLoader(&Config{DB: db, Debug: true})

	err = l.loadYml([]byte(yml), &ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	mock.ExpectBegin()

	mock.ExpectExec("^TRUNCATE TABLE \"table1\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"table2\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"table3\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	q := `^INSERT INTO "table1" AS table1_table_gonkey \("f1", "f2"\) VALUES ` +
		`\('value1', 'value2'\) ` +
		`RETURNING row_to_json\(table1_table_gonkey\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1\",\"f2\":\"value2\"}"),
		)

	q = `^INSERT INTO "table2" AS table2_table_gonkey \("f1", "f2"\) VALUES ` +
		`\('value2', 'value1'\) ` +
		`RETURNING row_to_json\(table2_table_gonkey\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value2\",\"f2\":\"value1\"}"),
		)

	q = `^INSERT INTO "table3" AS table3_table_gonkey \("f1", "f2"\) VALUES ` +
		`\('value1', 'value2'\) ` +
		`RETURNING row_to_json\(table3_table_gonkey\)$`

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
	yml := `
templates:
  baseTpl:
    f1: tplVal1
  ref3:
    $extend: baseTpl
    f2: tplVal2
tables:
  table1:
    - $name: ref1
      f1: value1
      f2: value2

  table2:
    - $name: ref2
      $extend: ref1
      f1: value1 overwritten
      f3: $eval("1" || "2" || 3 + 5)

  table3:
    - $extend: ref2
    - $extend: ref3
`

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctx := loadContext{
		refsDefinition: make(map[string]row),
		refsInserted:   make(map[string]row),
	}

	l := NewLoader(&Config{DB: db, Debug: true})

	err = l.loadYml([]byte(yml), &ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	mock.ExpectBegin()

	mock.ExpectExec("^TRUNCATE TABLE \"table1\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"table2\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("^TRUNCATE TABLE \"table3\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	q := `^INSERT INTO "table1" AS table1_table_gonkey \("f1", "f2"\) VALUES ` +
		`\('value1', 'value2'\) ` +
		`RETURNING row_to_json\(table1_table_gonkey\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1\",\"f2\":\"value2\"}"),
		)

	q = `^INSERT INTO "table2" AS table2_table_gonkey \("f1", "f2", "f3"\) VALUES ` +
		`\('value1 overwritten', 'value2', \(\"1\" \|\| \"2\" \|\| 3 \+ 5\)\) ` +
		`RETURNING row_to_json\(table2_table_gonkey\)$`

	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1 overwritten\",\"f2\":\"value2\",\"f3\":\"value3\"}"),
		)

	q = `^INSERT INTO "table3" AS table3_table_gonkey \("f1", "f2", "f3"\) VALUES ` +
		`\('value1 overwritten', 'value2', \(\"1\" \|\| \"2\" \|\| 3 \+ 5\)\), ` +
		`\('tplVal1', 'tplVal2', default\) ` +
		`RETURNING row_to_json\(table3_table_gonkey\)$`

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
