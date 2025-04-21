package multidb

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/fixtures/mysql"
	"github.com/lamoda/gonkey/fixtures/postgres"
	"github.com/lamoda/gonkey/models"
)

const (
	mysqlName        = "mysql"
	pgName           = "pg"
	fixturesLocation = "../testdata"
)

var idCounter int64

func TestMultiDBLoader(t *testing.T) {
	fixturesList := models.FixturesMultiDb{
		models.Fixture{
			DbName: mysqlName,
			Files: []string{
				"simple1.yaml",
			},
		},
		models.Fixture{
			DbName: pgName,
			Files: []string{
				"simple2.yaml",
			},
		},
	}

	mysqlDb, mysqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mysqlDb.Close() }()

	pgDb, pgMock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = pgDb.Close() }()

	l := New(map[string]fixtures.Loader{
		mysqlName: mysql.New(mysqlDb, fixturesLocation, true),
		pgName:    postgres.New(pgDb, fixturesLocation, true),
	})

	setMySQLMocks(t, mysqlMock)
	setPgMocks(t, pgMock)

	err = l.Load(fixturesList)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when loading fixtures", err)
	}

}

func setMySQLMocks(t *testing.T, mock sqlmock.Sqlmock) {
	mock.ExpectBegin()

	mock.ExpectExec("^TRUNCATE TABLE `table1`$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	expectMySQLInsert(t, mock,
		"table1",
		[]string{"field1", "field2"},
		"\\('value1', 2\\)",
		[]string{"value1", "2"},
	)

	mock.ExpectCommit()
}

func setPgMocks(t *testing.T, mock sqlmock.Sqlmock) {
	mock.ExpectBegin()

	mock.ExpectExec("^TRUNCATE TABLE \"public\".\"table2\" CASCADE$").
		WillReturnResult(sqlmock.NewResult(0, 0))

	q := `^INSERT INTO "public"."table2" AS row \("field3", "field4"\) VALUES ` +
		`\('value3', 4\) ` +
		`RETURNING row_to_json\(row\)$`
	mock.ExpectQuery(q).
		WillReturnRows(
			sqlmock.NewRows([]string{"json"}).
				AddRow("{\"f1\":\"value1\",\"f2\":\"value2\"}"),
		)

	mock.ExpectExec("^DO").
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectCommit()
}

func expectMySQLInsert(
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

func fieldsToDbStr(values []string) string {
	quotedVals := make([]string, len(values))

	for i, val := range values {
		quotedVals[i] = "`" + val + "`"
	}

	return "\\(" + strings.Join(quotedVals, ", ") + "\\)"
}
