package response_header

import (
	"fmt"
	"net/textproto"

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
	// test response headers with the expected headers
	expectedHeaders, ok := t.GetResponseHeaders(result.ResponseStatusCode)
	if !ok || len(expectedHeaders) == 0 {
		return nil, nil
	}

	var errs []error
	for k, v := range expectedHeaders {
		k = textproto.CanonicalMIMEHeaderKey(k)
		actualValues, ok := result.ResponseHeaders[k]
		if !ok {
			errs = append(errs, fmt.Errorf("response does not include expected header %s", k))
			continue
		}
		found := false
		for _, actualValue := range actualValues {
			e := compare.Compare(v, actualValue, compare.CompareParams{})
			if len(e) == 0 {
				found = true
			}
		}
		if !found {
			errs = append(errs, fmt.Errorf("response header %s value does not match expected %s", k, v))
		}
	}

	return errs, nil
}
