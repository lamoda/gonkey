package yaml_file

import (
	"github.com/lamoda/gonkey/models"
)

type Test struct {
	models.TestInterface

	TestDefinition

	Request         string
	Responses       map[int]string
	ResponseHeaders map[int]map[string][]string
	BeforeScript    string
	DbQuery         string
	DbResponse      []string
}

func (t *Test) ToQuery() string {
	return t.QueryParams
}

func (t *Test) GetMethod() string {
	return t.Method
}

func (t *Test) Path() string {
	return t.RequestURL
}

func (t *Test) ToJSON() ([]byte, error) {
	return []byte(t.Request), nil
}

func (t *Test) GetResponse(code int) (string, bool) {
	val, ok := t.Responses[code]
	return val, ok
}

func (t *Test) GetResponseHeaders(code int) (map[string][]string, bool) {
	val, ok := t.ResponseHeaders[code]
	return val, ok
}

func (t *Test) NeedsCheckingValues() bool {
	return !t.ComparisonParams.IgnoreValues
}

func (t *Test) GetName() string {
	return t.Name
}

func (t *Test) IgnoreArraysOrdering() bool {
	return t.ComparisonParams.IgnoreArraysOrdering
}

func (t *Test) DisallowExtraFields() bool {
	return t.ComparisonParams.DisallowExtraFields
}

func (t *Test) Fixtures() []string {
	return t.FixtureFiles
}

func (t *Test) ServiceMocks() map[string]interface{} {
	return t.MocksDefinition
}

func (t *Test) Pause() int {
	return t.PauseValue
}

func (t *Test) BeforeScriptPath() string {
	return t.BeforeScript
}

func (t *Test) BeforeScriptTimeout() int {
	return t.BeforeScriptParams.Timeout
}

func (t *Test) Cookies() map[string]string {
	return t.CookiesVal
}

func (t *Test) Headers() map[string]string {
	return t.HeadersVal
}

func (t *Test) DbQueryString() string {
	return t.DbQuery
}

func (t *Test) DbResponseJson() []string {
	return t.DbResponse
}
