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
		var prevCategory ErrorCategory

		for i, err := range testErrors {
			var checkErr *CheckError
			currentCategory := ErrorCategory("")

			if errors.As(err, &checkErr) {
				currentCategory = checkErr.GetCategory()
			}

			// Add empty line between different error categories
			if i > 0 && prevCategory != "" && currentCategory != prevCategory {
				errText += "\n"
			}

			errText = errText + err.Error() + "\n"
			prevCategory = currentCategory
		}

		return status, errors.New(errText)
	}

	return status, nil
}

// Passed returns true if test passed (false otherwise)
func (r *Result) Passed() bool {
	return len(r.Errors) == 0
}
