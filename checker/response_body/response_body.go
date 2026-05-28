package response_body

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/compare"
	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/xmlparsing"
)

type ResponseBodyChecker struct{}

func NewChecker() checker.CheckerInterface {
	return &ResponseBodyChecker{}
}

func (c *ResponseBodyChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errs []error
	var foundResponse bool
	if expectedBody, ok := t.GetResponse(result.ResponseStatusCode); ok {
		foundResponse = true
		switch {
		case strings.Contains(result.ResponseContentType, "json") && expectedBody != "":
			checkErrs, err := compareJsonBody(t, expectedBody, result)
			if err != nil {
				return nil, err
			}
			errs = append(errs, checkErrs...)
		case strings.Contains(result.ResponseContentType, "xml") && expectedBody != "":
			checkErrs, err := compareXmlBody(t, expectedBody, result)
			if err != nil {
				return nil, err
			}
			errs = append(errs, checkErrs...)
		default:
			compareErrs := compare.Compare(expectedBody, result.ResponseBody, compare.Params{})
			for _, err := range compareErrs {
				errs = append(errs, models.NewBodyErrorWithCause(err, "%s", err))
			}
		}

	}
	if !foundResponse {
		expectedCodes := getExpectedStatusCodes(t.GetResponses())
		err := models.NewStatusCodeError(expectedCodes[0], result.ResponseStatusCode)
		errs = append(errs, err)
	}

	return errs, nil
}

func getExpectedStatusCodes(responses map[int]string) []int {
	codes := make([]int, 0, len(responses))
	for code := range responses {
		codes = append(codes, code)
	}

	return codes
}

func compareJsonBody(t models.TestInterface, expectedBody string, result *models.Result) ([]error, error) {
	var expected interface{}
	if err := json.Unmarshal([]byte(expectedBody), &expected); err != nil {
		return nil, fmt.Errorf(
			"invalid JSON in response for test %s (status %d): %s",
			t.GetName(),
			result.ResponseStatusCode,
			err.Error(),
		)
	}

	var actual interface{}
	if err := json.Unmarshal([]byte(result.ResponseBody), &actual); err != nil {
		return []error{models.NewBodyErrorWithCause(err, "could not parse response")}, nil
	}

	params := compare.Params{
		IgnoreValues:         !t.NeedsCheckingValues(),
		IgnoreArraysOrdering: t.IgnoreArraysOrdering(),
		DisallowExtraFields:  t.DisallowExtraFields(),
	}

	compareErrs := compare.Compare(expected, actual, params)
	errs := make([]error, 0, len(compareErrs))
	for _, err := range compareErrs {
		errs = append(errs, models.NewBodyErrorWithCause(err, "%s", err))
	}

	return errs, nil
}

func compareXmlBody(t models.TestInterface, expectedBody string, result *models.Result) ([]error, error) {
	expected, err := xmlparsing.Parse(expectedBody)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid XML in response for test %s (status %d): %s",
			t.GetName(),
			result.ResponseStatusCode,
			err.Error(),
		)
	}

	actual, err := xmlparsing.Parse(result.ResponseBody)
	if err != nil {
		return []error{models.NewBodyErrorWithCause(err, "could not parse response")}, nil
	}
	params := compare.Params{
		IgnoreValues:         !t.NeedsCheckingValues(),
		IgnoreArraysOrdering: t.IgnoreArraysOrdering(),
		DisallowExtraFields:  t.DisallowExtraFields(),
	}

	compareErrs := compare.Compare(expected, actual, params)
	errs := make([]error, 0, len(compareErrs))
	for _, err := range compareErrs {
		errs = append(errs, models.NewBodyErrorWithCause(err, "%s", err))
	}

	return errs, nil
}
