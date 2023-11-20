package response_body

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/compare"
	"github.com/lamoda/gonkey/models"
)

type ResponseBodyChecker struct{}

func NewChecker() checker.CheckerInterface {
	return &ResponseBodyChecker{}
}

func (c *ResponseBodyChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errs []error
	var foundResponse bool
	// test response with the expected response body
	if expectedBody, ok := t.GetResponse(result.ResponseStatusCode); ok {
		foundResponse = true
		// is the response JSON document?
		if strings.Contains(result.ResponseContentType, "json") && expectedBody != "" {
			checkErrs, err := compareJsonBody(t, expectedBody, result)
			if err != nil {
				return nil, err
			}
			errs = append(errs, checkErrs...)
		} else {
			// compare bodies as leaf nodes
			errs = append(errs, compare.Compare(expectedBody, result.ResponseBody, compare.Params{})...)
		}
	}
	if !foundResponse {
		err := fmt.Errorf("server responded with status %d", result.ResponseStatusCode)
		errs = append(errs, err)
	}

	return errs, nil
}

func compareJsonBody(t models.TestInterface, expectedBody string, result *models.Result) ([]error, error) {
	// decode expected body
	var expected interface{}
	if err := json.Unmarshal([]byte(expectedBody), &expected); err != nil {
		return nil, fmt.Errorf(
			"invalid JSON in response for test %s (status %d): %s",
			t.GetName(),
			result.ResponseStatusCode,
			err.Error(),
		)
	}

	// decode actual body
	var actual interface{}
	if err := json.Unmarshal([]byte(result.ResponseBody), &actual); err != nil {
		return []error{errors.New("could not parse response")}, nil
	}

	params := compare.Params{
		IgnoreValues:         !t.NeedsCheckingValues(),
		IgnoreArraysOrdering: t.IgnoreArraysOrdering(),
		DisallowExtraFields:  t.DisallowExtraFields(),
	}

	return compare.Compare(expected, actual, params), nil
}
