package variables

import (
	"regexp"

	"github.com/lamoda/gonkey/models"
)

type Variables struct {
	variables variables
}

type variables map[string]*Variable

var variableRx = regexp.MustCompile(`{{\s*\$(\w+)\s*}}`)

func New() *Variables {
	return &Variables{
		variables: make(variables),
	}
}

// Load adds new variables and replaces values of existing
func (vs *Variables) Load(variables map[string]string) {
	for n, v := range variables {
		variable := NewVariable(n, v)

		vs.variables[n] = variable
	}
}

// Load adds new variables and replaces values of existing
func (vs *Variables) Set(name, value string) {
	v := NewVariable(name, value)

	vs.variables[name] = v
}

func (vs *Variables) Apply(t models.TestInterface) models.TestInterface {

	newTest := t.Clone()

	if vs == nil {
		return newTest
	}

	newTest.SetQuery(vs.perform(newTest.ToQuery()))
	newTest.SetMethod(vs.perform(newTest.GetMethod()))
	newTest.SetPath(vs.perform(newTest.Path()))
	newTest.SetRequest(vs.perform(newTest.GetRequest()))

	newTest.SetResponses(vs.performResponses(newTest.GetResponses()))
	newTest.SetHeaders(vs.performHeaders(newTest.Headers()))

	return newTest
}

// Merge adds given variables to set or overrides existed
func (vs *Variables) Merge(vars *Variables) {
	for k, v := range vars.variables {
		vs.variables[k] = v
	}
}

func (vs *Variables) Len() int {
	return len(vs.variables)
}

func usedVariables(str string) (res []string) {
	matches := variableRx.FindAllStringSubmatch(str, -1)
	for _, match := range matches {
		res = append(res, match[1])
	}

	return res
}

// perform replaces all variables in str to their values
// and returns result string
func (vs *Variables) perform(str string) string {

	varNames := usedVariables(str)

	for _, k := range varNames {
		if v := vs.get(k); v != nil {
			str = v.Perform(str)
		}
	}

	return str
}

func (vs *Variables) get(name string) *Variable {

	v := vs.variables[name]
	if v == nil {
		v = NewFromEnvironment(name)
	}

	return v
}

func (vs *Variables) performHeaders(headers map[string]string) map[string]string {

	res := make(map[string]string)

	for k, v := range headers {
		res[k] = vs.perform(v)
	}
	return res
}

func (vs *Variables) performResponses(responses map[int]string) map[int]string {

	res := make(map[int]string)

	for k, v := range responses {
		res[k] = vs.perform(v)
	}
	return res
}

func (vs *Variables) Add(v *Variable) *Variables {
	vs.variables[v.name] = v

	return vs
}
