package response_header

import (
	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/compare"
	"github.com/lamoda/gonkey/models"
)

type ResponseHeaderChecker struct {
	checker.CheckerInterface
}

func NewChecker() checker.CheckerInterface {
	return &ResponseHeaderChecker{}
}

func (c *ResponseHeaderChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errs []error

	// test response headers with the expected headers
	if expectedHeader, ok := t.GetResponseHeaders(result.ResponseStatusCode); ok {
		if len(expectedHeader) > 0 {
			errs = append(errs, compare.Compare(expectedHeader, result.ResponseHeaders, compare.CompareParams{})...)
		}
	}

	return errs, nil
}
