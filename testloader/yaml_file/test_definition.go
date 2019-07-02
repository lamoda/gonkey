package yaml_file

type TestDefinition struct {
	Name               string                 `json:"name" yaml:"name"`
	Method             string                 `json:"method" yaml:"method"`
	RequestURL         string                 `json:"path" yaml:"path"`
	QueryParams        string                 `json:"query" yaml:"query"`
	RequestTmpl        string                 `json:"request" yaml:"request"`
	ResponseTmpls      map[int]string         `json:"response" yaml:"response"`
	BeforeScriptParams beforeScriptParams     `json:"beforeScript" yaml:"beforeScript"`
	HeadersVal         map[string]string      `json:"headers" yaml:"headers"`
	CookiesVal         map[string]string      `json:"cookies" yaml:"cookies"`
	Cases              []CaseData             `json:"cases" yaml:"cases"`
	ComparisonParams   comparisonParams       `json:"comparisonParams" yaml:"comparisonParams"`
	FixtureFiles       []string               `json:"fixtures" yaml:"fixtures"`
	MocksDefinition    map[string]interface{} `json:"mocks" yaml:"mocks"`
	PauseValue         int                    `json:"pause" yaml:"pause"`
	DbQueryTmpl        string                 `json:"dbQuery" yaml:"dbQuery"`
	DbResponseTmpl     []string               `json:"dbResponse" yaml:"dbResponse"`
}

type CaseData struct {
	RequestArgs      map[string]interface{}         `json:"requestArgs" yaml:"requestArgs"`
	ResponseArgs     map[int]map[string]interface{} `json:"responseArgs" yaml:"responseArgs"`
	BeforeScriptArgs map[string]interface{}         `json:"beforeScriptArgs" yaml:"beforeScriptArgs"`
	DbQueryArgs      map[string]interface{}         `json:"dbQueryArgs" yaml:"dbQueryArgs"`
	DbResponseArgs   map[string]interface{}         `json:"dbResponseArgs" yaml:"dbResponseArgs"`
	DbResponse       []string                       `json:"dbResponse" yaml:"dbResponse"`
}

type comparisonParams struct {
	IgnoreValues         bool `json:"ignoreValues" yaml:"ignoreValues"`
	IgnoreArraysOrdering bool `json:"ignoreArraysOrdering" yaml:"ignoreArraysOrdering"`
	DisallowExtraFields  bool `json:"disallowExtraFields" yaml:"disallowExtraFields"`
}

type beforeScriptParams struct {
	PathTmpl string `json:"path" yaml:"path"`
	Timeout  int    `json:"timeout" yaml:"timeout"`
}
