package variables

import "github.com/lamoda/gonkey/models"

type Variables struct {
	variables variables
}

type variables map[string]*Variable

func New() *Variables {
	return &Variables{
		variables: make(variables),
	}
}

// Load adds new variables and replaces values of existing
func (vs *Variables) Load(variables map[string]string) error {
	for n, v := range variables {
		variable, err := NewVariable(n, v)
		if err != nil {
			return err
		}
		vs.variables[n] = variable
	}

	return nil
}

func (vs *Variables) Apply(t models.TestInterface) models.TestInterface {

	res := t.Clone()

	if vs == nil {
		return res
	}

	res.SetQuery(vs.perform(res.ToQuery()))
	res.SetMethod(vs.perform(res.GetMethod()))
	res.SetPath(vs.perform(res.Path()))
	res.SetRequest(vs.perform(res.GetRequest()))

	res.SetResponses(vs.performResponses(res.GetResponses()))
	res.SetHeaders(vs.performHeaders(res.Headers()))

	return res
}

// perform replaces all variables in str to their values
// and returns result string
func (vs *Variables) perform(str string) string {

	for _, v := range vs.variables {
		str = v.Perform(str)
	}

	return str
}

func (vs *Variables) Len() int {
	return len(vs.variables)
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
