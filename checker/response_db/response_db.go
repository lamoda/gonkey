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
	var checkErrors []error
	if t.IgnoreDbOrdering() {
		checkErrors, err = compareDbRespWithoutOrdering(t.DbResponseJson(), result.DbResponse, t.GetName())
	} else {
		checkErrors, err = compareDbResp(t.DbResponseJson(), result.DbResponse, t.GetName(), result.DbQuery)
	}
	if err != nil {
		return nil, err
	}
	errors = append(errors, checkErrors...)

	return errors, nil
}

func compareDbRespWithoutOrdering(expected, actual []string, testName string) ([]error, error) {
	var errors []error
	var actualJsons []interface{}
	var expectedJsons []interface{}

	// gather expected and actual rows
	for i, row := range expected {
		// decode expected row
		var expectedJson interface{}
		if err := json.Unmarshal([]byte(row), &expectedJson); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the expected DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				testName,
				i,
				row,
				err.Error(),
			)
		}
		expectedJsons = append(expectedJsons, expectedJson)
		// decode actual row
		var actualJson interface{}
		if err := json.Unmarshal([]byte(actual[i]), &actualJson); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the actual DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				testName,
				i,
				actual[i],
				err.Error(),
			)
		}
		actualJsons = append(actualJsons, actualJson)
	}

	remove := func(array []interface{}, i int) []interface{} {
		array[i] = array[len(array)-1]
		return array[:len(array)-1]
	}

	// compare actual and expected rows
	for _, actualRow := range actualJsons {
		for i, expectedRow := range expectedJsons {
			if diff := pretty.Compare(expectedRow, actualRow); diff == "" {
				expectedJsons = remove(expectedJsons, i)
				break
			}
		}
	}

	if len(expectedJsons) > 0 {
		errorString := "missing expected items in database:"

		for _, expectedRow := range expectedJsons {
			expectedRowJson, _ := json.Marshal(expectedRow)
			errorString += fmt.Sprintf("\n - %s", color.CyanString("%s", expectedRowJson))
		}

		errors = append(errors, fmt.Errorf(errorString))
	}

	return errors, nil
}

func compareDbResp(expected, actual []string, testName string, query interface{}) ([]error, error) {
	var errors []error
	var actualJson interface{}
	var expectedJson interface{}

	for i, row := range expected {
		// decode expected row
		if err := json.Unmarshal([]byte(row), &expectedJson); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the expected DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				testName,
				i,
				row,
				err.Error(),
			)
		}
		// decode actual row
		if err := json.Unmarshal([]byte(actual[i]), &actualJson); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the actual DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				testName,
				i,
				actual[i],
				err.Error(),
			)
		}

		// compare responses row as jsons
		if err := compareDbResponseRow(expectedJson, actualJson, query); err != nil {
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
