package response_db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/compare"
	"github.com/lamoda/gonkey/models"

	"github.com/fatih/color"
	"github.com/kylelemons/godebug/pretty"
)

type ResponseDbChecker struct {
	db *sql.DB
}

func NewChecker(dbConnect *sql.DB) checker.CheckerInterface {
	return &ResponseDbChecker{
		db: dbConnect,
	}
}

func (c *ResponseDbChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errors []error
	errs, err := c.check(t.GetName(), t.IgnoreDbOrdering(), t, result)
	if err != nil {
		return nil, err
	}
	errors = append(errors, errs...)

	for _, dbCheck := range t.GetDatabaseChecks() {
		errs, err := c.check(t.GetName(), t.IgnoreDbOrdering(), dbCheck, result)
		if err != nil {
			return nil, err
		}
		errors = append(errors, errs...)
	}

	return errors, nil
}

func (c *ResponseDbChecker) check(
	testName string,
	ignoreOrdering bool,
	t models.DatabaseCheck,
	result *models.Result,
) ([]error, error) {
	var errors []error

	// don't check if there are no data for db test
	if t.DbQueryString() == "" && t.DbResponseJson() == nil {
		return errors, nil
	}

	// check expected db query exist
	if t.DbQueryString() == "" {
		return nil, fmt.Errorf("DB query not found for test \"%s\"", testName)
	}

	// check expected response exist
	if t.DbResponseJson() == nil {
		return nil, fmt.Errorf("expected DB response not found for test \"%s\"", testName)
	}

	// get DB response
	actualDbResponse, err := newQuery(t.DbQueryString(), c.db)
	if err != nil {
		return nil, err
	}

	result.DatabaseResult = append(
		result.DatabaseResult,
		models.DatabaseResult{Query: t.DbQueryString(), Response: actualDbResponse},
	)

	// compare responses length
	err = compareDbResponseLength(t.DbResponseJson(), actualDbResponse, t.DbQueryString())
	if err != nil {
		return append(errors, err), nil
	}
	// compare responses as json lists
	expectedItems, err := toJSONArray(t.DbResponseJson(), "expected", testName)
	if err != nil {
		return nil, err
	}
	actualItems, err := toJSONArray(actualDbResponse, "actual", testName)
	if err != nil {
		return nil, err
	}

	errs := compare.Compare(expectedItems, actualItems, compare.Params{
		IgnoreArraysOrdering: ignoreOrdering,
	})

	errors = append(errors, errs...)

	return errors, nil
}

func toJSONArray(items []string, qual, testName string) ([]interface{}, error) {
	itemJSONs := make([]interface{}, 0, len(items))
	for i, row := range items {
		var itemJSON interface{}
		if err := json.Unmarshal([]byte(row), &itemJSON); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the %s DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				qual,
				testName,
				i,
				row,
				err.Error(),
			)
		}
		itemJSONs = append(itemJSONs, itemJSON)
	}

	return itemJSONs, nil
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
