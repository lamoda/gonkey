package models

// Result of test execution
type Result struct {
	Path                string // TODO: remove
	Query               string // TODO: remove
	RequestBody         string
	ResponseStatusCode  int
	ResponseStatus      string
	ResponseContentType string
	ResponseBody        string
	Errors              []error
	Test                TestInterface
}

// Passed returns true if test passed (false otherwise)
func (r *Result) Passed() bool {
	return len(r.Errors) == 0
}
