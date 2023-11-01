package models

import "errors"

type DatabaseResult struct {
	Query    string
	Response []string
}

// Result of test execution
type Result struct {
	Path                string // TODO: remove
	Query               string // TODO: remove
	RequestBody         string
	ResponseStatusCode  int
	ResponseStatus      string
	ResponseContentType string
	ResponseBody        string
	ResponseHeaders     map[string][]string
	Errors              []error
	Test                TestInterface
	DatabaseResult      []DatabaseResult
}

func allureStatus(status string) bool {
	switch status {
	case "passed", "failed", "broken", "skipped":
		return true
	default:
		return false
	}
}

func notRunnedStatus(status string) bool {
	switch status {
	case "broken", "skipped":
		return true
	default:
		return false
	}
}

func (r *Result) AllureStatus() (string, error) {
	testStatus := r.Test.GetStatus()
	if testStatus != "" && allureStatus(testStatus) && notRunnedStatus(testStatus) {
		return testStatus, nil
	}

	var (
		status     = "passed"
		testErrors []error
	)

	if len(r.Errors) != 0 {
		status = "failed"
		testErrors = r.Errors
	}

	if len(testErrors) != 0 {
		errText := ""
		for _, err := range testErrors {
			errText = errText + err.Error() + "\n"
		}

		return status, errors.New(errText)
	}

	return status, nil
}

// Passed returns true if test passed (false otherwise)
func (r *Result) Passed() bool {
	return len(r.Errors) == 0
}
