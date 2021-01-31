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
	GetResponseHeaders(code int) (map[string]string, bool)
	GetName() string
	Fixtures() []string
	ServiceMocks() map[string]interface{}
	Pause() int
	BeforeScriptPath() string
	BeforeScriptTimeout() int
	Cookies() map[string]string
	Headers() map[string]string
	ContentType() string
	GetForm() *Form
	DbQueryString() string
	DbResponseJson() []string
	GetVariables() map[string]string
	GetVariablesToSet() map[int]map[string]string
	ResponseTransformScripts() map[int][]string
	ResponseTransformTimeout() int

	// setters
	SetQuery(string)
	SetMethod(string)
	SetPath(string)
	SetRequest(string)
	SetForm(form *Form)
	SetResponses(map[int]string)
	SetHeaders(map[string]string)

	// comparison properties
	NeedsCheckingValues() bool
	IgnoreArraysOrdering() bool
	DisallowExtraFields() bool

	// Clone returns copy of current object
	Clone() TestInterface
}

// TODO: add support for form fields
type Form struct {
	Files map[string]string `json:"files" yaml:"files"`
}

type Summary struct {
	Success bool
	Failed  int
	Total   int
}
