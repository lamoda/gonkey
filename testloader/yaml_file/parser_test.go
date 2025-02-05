package yaml_file

import (
	"fmt"
	"os"
	"testing"
)

var testsYAMLData = `
- method: POST
  path: /jsonrpc/v2/orders.nr
  request:
    '{
      "jsonrpc": "2.0",
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "method": "orders.nr",
      "params": [
        {
          "amount": 1,
          "prefix": "ru"
        }
      ]
    }'
  response:
    200:
      '{
         "result": [
           {
             "nr": "number",
             "prefix": "ru",
             "vc": "vc"
           }
         ],
         "id": "550e8400-e29b-41d4-a716-446655440000",
         "jsonrpc": "2.0"
       }'
  cases:
    - requestArgs:
        foo: 'Hello world'
        bar: 42
      responseArgs:
        200:
          foo: 'Hello world'
          bar: 42
    - requestArgs:
        foo: 'Hello world'
        bar: 42
      responseArgs:
        200:
          foo: 'Hello world'
          bar: 42
      variables:
        newVar: some_value
`

func TestParseTestsWithCases(t *testing.T) {
	tmpfile, err := os.CreateTemp("../..", "tmpfile_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := fmt.Fprint(tmpfile, testsYAMLData); err != nil {
		t.Fatal(err)
	}

	tests, err := parseTestDefinitionFile(tmpfile.Name())
	if err != nil {
		t.Error(err)
	}
	if len(tests) != 2 {
		t.Errorf("wait len(tests) == 2, got len(tests) == %d", len(tests))
	}
}
