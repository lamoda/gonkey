package yaml_file

import (
	"strings"

	"github.com/lamoda/gonkey/models"
)

type dbCheck struct {
	query    string
	response []string
}

func (c *dbCheck) DbQueryString() string        { return c.query }
func (c *dbCheck) DbResponseJson() []string     { return c.response }
func (c *dbCheck) SetDbQueryString(q string)    { c.query = q }
func (c *dbCheck) SetDbResponseJson(r []string) { c.response = r }

type Test struct {
	TestDefinition

	Filename string

	Request            string
	Responses          map[int]string
	ResponseHeaders    map[int]map[string]string
	BeforeScript       string
	AfterRequestScript string
	DbQuery            string
	DbResponse         []string

	CombinedVariables map[string]string

	DbChecks []models.DatabaseCheck
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

func (t *Test) GetRequest() string {
	return t.Request
}

func (t *Test) ToJSON() ([]byte, error) {
	return []byte(t.Request), nil
}

func (t *Test) GetResponses() map[int]string {
	return t.Responses
}

func (t *Test) GetResponse(code int) (string, bool) {
	val, ok := t.Responses[code]

	return val, ok
}

func (t *Test) GetResponseHeaders(code int) (map[string]string, bool) {
	val, ok := t.ResponseHeaders[code]

	return val, ok
}

func (t *Test) NeedsCheckingValues() bool {
	return !t.ComparisonParams.IgnoreValues
}

func (t *Test) GetName() string {
	return t.Name
}

func (t *Test) GetDescription() string {
	return t.Description
}

func (t *Test) GetStatus() string {
	return t.Status
}

func (t *Test) IgnoreArraysOrdering() bool {
	return t.ComparisonParams.IgnoreArraysOrdering
}

func (t *Test) DisallowExtraFields() bool {
	return t.ComparisonParams.DisallowExtraFields
}

func (t *Test) IgnoreDbOrdering() bool {
	return t.ComparisonParams.IgnoreDbOrdering
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

func (t *Test) AfterRequestScriptPath() string {
	return t.AfterRequestScript
}

func (t *Test) AfterRequestScriptTimeout() int {
	return t.AfterRequestScriptParams.Timeout
}

func (t *Test) Cookies() map[string]string {
	return t.CookiesVal
}

func (t *Test) Headers() map[string]string {
	return t.HeadersVal
}

// TODO: it might make sense to do support of case-insensitive checking
func (t *Test) ContentType() string {
	return t.HeadersVal["Content-Type"]
}

func (t *Test) DbQueryString() string {
	return t.DbQuery
}

func (t *Test) DbResponseJson() []string {
	return t.DbResponse
}

func (t *Test) GetDatabaseChecks() []models.DatabaseCheck       { return t.DbChecks }
func (t *Test) SetDatabaseChecks(checks []models.DatabaseCheck) { t.DbChecks = checks }

func (t *Test) GetVariables() map[string]string {
	return t.Variables
}

func (t *Test) GetCombinedVariables() map[string]string {
	return t.CombinedVariables
}

func (t *Test) GetForm() *models.Form {
	return t.Form
}

func (t *Test) GetVariablesToSet() map[int]map[string]string {
	return t.VariablesToSet
}

func (t *Test) GetFileName() string {
	return t.Filename
}

func (t *Test) Clone() models.TestInterface {
	res := *t

	return &res
}

func (t *Test) SetQuery(val string) {
	var query strings.Builder
	query.Grow(len(val) + 1)
	if val != "" && val[0] != '?' {
		query.WriteString("?")
	}
	query.WriteString(val)
	t.QueryParams = query.String()
}

func (t *Test) SetMethod(val string) {
	t.Method = val
}

func (t *Test) SetPath(val string) {
	t.RequestURL = val
}

func (t *Test) SetRequest(val string) {
	t.Request = val
}

func (t *Test) SetForm(val *models.Form) {
	t.Form = val
}

func (t *Test) SetResponses(val map[int]string) {
	t.Responses = val
}

func (t *Test) SetHeaders(val map[string]string) {
	t.HeadersVal = val
}

func (t *Test) SetDbQueryString(query string) {
	t.DbQuery = query
}

func (t *Test) SetDbResponseJson(responses []string) {
	t.DbResponse = responses
}

func (t *Test) SetStatus(status string) {
	t.Status = status
}
