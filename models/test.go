package models

// Common Test interface
type TestInterface interface {
	ToQuery() string
	GetRequest() string
	ToJSON() ([]byte, error)
	GetMethod() string
	Path() string
	GetResponses() map[int]string
	GetResponse(code int) (string, bool)
	GetName() string
	Fixtures() []string
	ServiceMocks() map[string]interface{}
	Pause() int
	BeforeScriptPath() string
	BeforeScriptTimeout() int
	Cookies() map[string]string
	Headers() map[string]string
	DbQueryString() string
	DbResponseJson() []string
	GetVariables() map[string]string

	// setters
	SetQuery(string)
	SetMethod(string)
	SetPath(string)
	SetRequest(string)
	SetResponses(map[int]string)
	SetHeaders(map[string]string)

	// comparison properties
	NeedsCheckingValues() bool
	IgnoreArraysOrdering() bool
	DisallowExtraFields() bool

	// Clone returns copy of current object
	Clone() TestInterface
}

type Summary struct {
	Success bool
	Failed  int
	Total   int
}
