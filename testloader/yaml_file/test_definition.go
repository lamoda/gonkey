package yaml_file

type TestDefinition struct {
	Name               string                    `json:"name" yaml:"name"`
	Variables          map[string]string         `json:"variables" yaml:"variables"`
	VariablesToSet     VariablesToSet            `json:"variables_to_set" yaml:"variables_to_set"`
	Method             string                    `json:"method" yaml:"method"`
	RequestURL         string                    `json:"path" yaml:"path"`
	QueryParams        string                    `json:"query" yaml:"query"`
	RequestTmpl        string                    `json:"request" yaml:"request"`
	ResponseTmpls      map[int]string            `json:"response" yaml:"response"`
	ResponseHeaders    map[int]map[string]string `json:"responseHeaders" yaml:"responseHeaders"`
	BeforeScriptParams beforeScriptParams        `json:"beforeScript" yaml:"beforeScript"`
	HeadersVal         map[string]string         `json:"headers" yaml:"headers"`
	CookiesVal         map[string]string         `json:"cookies" yaml:"cookies"`
	Cases              []CaseData                `json:"cases" yaml:"cases"`
	ComparisonParams   comparisonParams          `json:"comparisonParams" yaml:"comparisonParams"`
	FixtureFiles       []string                  `json:"fixtures" yaml:"fixtures"`
	MocksDefinition    map[string]interface{}    `json:"mocks" yaml:"mocks"`
	PauseValue         int                       `json:"pause" yaml:"pause"`
	DbQueryTmpl        string                    `json:"dbQuery" yaml:"dbQuery"`
	DbResponseTmpl     []string                  `json:"dbResponse" yaml:"dbResponse"`
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

type VariablesToSet map[int]map[string]string

/*
There can be two types of data in yaml-file:
1) JSON-paths:
	VariablesToSet:
		<code1>:
			<varName1>: <JSON_Path1>
			<varName2>: <JSON_Path2>
2) Plain text:
	 VariablesToSet:
		<code1>: <varName1>
		<code2>: <varName2>
		...
   In this case we unmarshall values to format similar to JSON-paths format with empty paths:
	 VariablesToSet:
		<code1>:
			<varName1>: ""
		<code2>:
			<varName2>: ""
*/
func (v *VariablesToSet) UnmarshalYAML(unmarshal func(interface{}) error) error {

	res := make(map[int]map[string]string)

	// try to unmarshall as plaint text
	var plain map[int]string
	if err := unmarshal(&plain); err == nil {

		for code, varName := range plain {
			res[code] = map[string]string{
				varName: "",
			}
		}

		*v = res
		return nil
	}

	// json-paths
	if err := unmarshal(&res); err != nil {
		return err
	}

	*v = res
	return nil
}
