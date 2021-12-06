package yaml_file

import (
	"reflect"
	"testing"
)

func TestNewTestWithCases(t *testing.T) {
	data := TestDefinition{
		RequestTmpl: `{"foo": "bar", "hello": {{ .hello }} }`,
		ResponseTmpls: map[int]string{
			200: `{"foo": "bar", "hello": {{ .hello }} }`,
			400: `{"foo": "bar", "hello": {{ .hello }} }`,
		},
		ResponseHeaders: map[int]map[string]string{
			200: {
				"hello": "world",
				"say":   "hello",
			},
			400: {
				"hello": "world",
				"foo":   "bar",
			},
		},
		Cases: []CaseData{
			{
				RequestArgs: map[string]interface{}{
					"hello": `"world"`,
				},
				ResponseArgs: map[int]map[string]interface{}{
					200: {
						"hello": "world",
					},
					400: {
						"hello": "world",
					},
				},
			},
			{
				RequestArgs: map[string]interface{}{
					"hello": `"world2"`,
				},
				ResponseArgs: map[int]map[string]interface{}{
					200: {
						"hello": "world2",
					},
					400: {
						"hello": "world2",
					},
				},
			},
		},
	}

	tests, err := makeTestFromDefinition("cases/example.yaml", data)

	if err != nil {
		t.Fatal(err)
	}
	if len(tests) != 2 {
		t.Errorf("wait len(tests) == 2, got len(tests) == %d", len(tests))
	}

	reqData, err := tests[0].ToJSON()
	if !reflect.DeepEqual(reqData, []byte(`{"foo": "bar", "hello": "world" }`)) {
		t.Errorf("want request %s, got %s", `{"foo": "bar", "hello": "world" }`, reqData)
	}

	filename := tests[0].GetFileName()
	if !reflect.DeepEqual(filename, "cases/example.yaml") {
		t.Errorf("want filename %s, got %s", "cases/example.yaml", filename)
	}

	reqData, err = tests[1].ToJSON()
	if !reflect.DeepEqual(reqData, []byte(`{"foo": "bar", "hello": "world2" }`)) {
		t.Errorf("want request %s, got %s", `{"foo": "bar", "hello": "world2" }`, reqData)
	}

	filename = tests[1].GetFileName()
	if !reflect.DeepEqual(filename, "cases/example.yaml") {
		t.Errorf("want filename %s, got %s", "cases/example.yaml", filename)
	}
}
