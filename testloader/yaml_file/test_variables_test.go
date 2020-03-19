package yaml_file

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/variables"
)

const requestOriginal = `{"reqParam": "{{ $reqParam }}"}`
const requestApplied = `{"reqParam": "reqParam_value"}`

func TestParseTestsWithVariables(t *testing.T) {

	tests, err := parseTestDefinitionFile("testdata/variables.yaml")
	if err != nil {
		t.Error(err)
	}

	testOriginal := &tests[0]

	vars := variables.New()
	vars.Load(testOriginal.GetVariables())
	testApplied := vars.Apply(testOriginal)

	// check that original test is not changed
	checkOriginal(t, testOriginal)

	checkApplied(t, testApplied)
}

func checkOriginal(t *testing.T, test models.TestInterface) {

	t.Helper()

	req, err := test.ToJSON()
	assert.NoError(t, err)
	assert.Equal(t, requestOriginal, string(req))

	assert.Equal(t, "{{ $method }}", test.GetMethod())
	assert.Equal(t, "/some/path/{{ $pathPart }}", test.Path())
	assert.Equal(t, "{{ $query }}", test.ToQuery())
	assert.Equal(t, map[string]string{"header1": "{{ $header }}"}, test.Headers())

	resp, ok := test.GetResponse(200)
	assert.True(t, ok)
	assert.Equal(t, "{{ $resp }}", resp)

	resp, ok = test.GetResponse(404)
	assert.True(t, ok)
	assert.Equal(t, "{{ $respRx }}", resp)
}

func checkApplied(t *testing.T, test models.TestInterface) {

	t.Helper()

	req, err := test.ToJSON()
	assert.NoError(t, err)
	assert.Equal(t, requestApplied, string(req))

	assert.Equal(t, "POST", test.GetMethod())
	assert.Equal(t, "/some/path/part_of_path", test.Path())
	assert.Equal(t, "query_val", test.ToQuery())
	assert.Equal(t, map[string]string{"header1": "header_val"}, test.Headers())

	resp, ok := test.GetResponse(200)
	assert.True(t, ok)
	assert.Equal(t, "resp_val", resp)

	resp, ok = test.GetResponse(404)
	assert.True(t, ok)
	assert.Equal(t, "$matchRegexp(^[0-9.]+$)", resp)
}
