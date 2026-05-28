package response_db

import (
	"database/sql"
	"fmt"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/models"
)

type ResponseMultiDbChecker struct {
	checkers map[string]MultiDBCheckerInstance
}

type MultiDBCheckerInstance interface {
	check(testName string, ignoreOrdering bool, t models.DatabaseCheck, result *models.Result, queryIndex int) ([]error, error)
}

func NewMultiDbChecker(dbMap map[string]*sql.DB) checker.CheckerInterface {
	checkers := make(map[string]MultiDBCheckerInstance, len(dbMap))
	for dbName, conn := range dbMap {
		checkers[dbName] = &ResponseDbChecker{
			db: conn,
		}
	}

	return &ResponseMultiDbChecker{checkers: checkers}
}

func (c *ResponseMultiDbChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errors []error

	for _, dbCheck := range t.GetDatabaseChecks() {
		dbChecker, ok := c.checkers[dbCheck.DbNameString()]
		if !ok || dbChecker == nil {
			return nil, fmt.Errorf("DB name %s not mapped not found for test \"%s\"", dbCheck.DbNameString(), t.GetName())
		}

		queryIndex := len(result.DatabaseResult)
		errs, err := dbChecker.check(t.GetName(), t.IgnoreDbOrdering(), dbCheck, result, queryIndex)
		if err != nil {
			return nil, err
		}

		errors = append(errors, errs...)
	}

	return errors, nil
}
