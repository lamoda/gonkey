package response_db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/models"

	"github.com/fatih/color"
	"github.com/kylelemons/godebug/pretty"
)

type ResponseDbChecker struct {
	checker.CheckerInterface

	db *sql.DB
}

func NewChecker(dbConnect *sql.DB) checker.CheckerInterface {
	return &ResponseDbChecker{
		db: dbConnect,
	}
}

func (c *ResponseDbChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errors []error

	// don't check if there are no data for db test
	if t.DbQueryString() == "" && t.DbResponseJson() == nil {
		return errors, nil
	}

	// check expected db query exist
	if t.DbQueryString() == "" {
		return nil, fmt.Errorf(
			"DB query not found for test \"%s\"",
			t.GetName(),
		)
	}

	// check expected response exist
	if t.DbResponseJson() == nil {
		return nil, fmt.Errorf(
			"expected DB response not found for test \"%s\"",
			t.GetName(),
		)
	}

	// get DB response
	actualDbResponse, err := newQuery(t.DbQueryString(), c.db)
	if err != nil {
		return nil, err
	}
	result.DbQuery = t.DbQueryString()
	result.DbResponse = actualDbResponse

	// compare responses length
	if err := compareDbResponseLength(t.DbResponseJson(), result.DbResponse, result.DbQuery); err != nil {
		errors = append(errors, err)
		return errors, nil
	}
	// compare responses as json lists
	checkErrors, err := compareDbResp(t, result)
	if err != nil {
		return nil, err
	}
	errors = append(errors, checkErrors...)

	return errors, nil
}

func compareDbResp(t models.TestInterface, result *models.Result) ([]error, error) {
	var errors []error
	var actualJson interface{}
	var expectedJson interface{}

	for i, row := range t.DbResponseJson() {
		// decode expected row
		if err := json.Unmarshal([]byte(row), &expectedJson); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the expected DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				t.GetName(),
				i,
				row,
				err.Error(),
			)
		}
		// decode actual row
		if err := json.Unmarshal([]byte(result.DbResponse[i]), &actualJson); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the actual DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				t.GetName(),
				i,
				result.DbResponse[i],
				err.Error(),
			)
		}

		// compare responses row as jsons
		if err := compareDbResponseRow(expectedJson, actualJson, result.DbQuery); err != nil {
			errors = append(errors, err)
		}
	}

	return errors, nil
}

func compareDbResponseRow(expected, actual, query interface{}) error {
	var err error

	if diff := pretty.Compare(expected, actual); diff != "" {
		err = fmt.Errorf(
			"items in database do not match (-expected: +actual):\n     test query:\n%s\n    result diff:\n%s",
			color.CyanString("%v", query),
			color.CyanString("%v", diff),
		)
	}
	return err
}

func compareDbResponseLength(expected, actual []string, query interface{}) error {
	var err error

	if len(expected) != len(actual) {
		err = fmt.Errorf(
			"quantity of items in database do not match (-expected: %s +actual: %s)\n     test query:\n%s\n    result diff:\n%s",
			color.CyanString("%v", len(expected)),
			color.CyanString("%v", len(actual)),
			color.CyanString("%v", query),
			color.CyanString("%v", pretty.Compare(expected, actual)),
		)
	}
	return err
}

func newQuery(dbQuery string, db *sql.DB) ([]string, error) {

	var dbResponse []string
	var jsonString string

	if idx := strings.IndexByte(dbQuery, ';'); idx >= 0 {
		dbQuery = dbQuery[:idx]
	}

	rows, err := db.Query(fmt.Sprintf("SELECT row_to_json(rows) FROM (%s) rows;", dbQuery))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&jsonString)
		if err != nil {
			return nil, err
		}
		dbResponse = append(dbResponse, jsonString)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return dbResponse, nil
}

