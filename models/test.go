package models

// Common Test interface
type TestInterface interface {
	ToQuery() string
	ToJSON() ([]byte, error)
	GetMethod() string
	Path() string
	GetResponse(code int) (string, bool)
	GetName() string
	Fixtures() []string
	ServiceMocks() map[string]interface{}
	Pause() int
	BeforeScriptPath() string
	BeforeScriptTimeout() int
	Cookies() map[string]string
	Headers() map[string]string

	// comparison properties
	NeedsCheckingValues() bool
	IgnoreArraysOrdering() bool
	DisallowExtraFields() bool
}

type Summary struct {
	Success bool
	Failed  int
	Total   int
}
