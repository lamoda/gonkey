package runner

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

const (
	body = "bla"
)

func Test_PreviousResult(t *testing.T) {
	srv := testServer()
	defer srv.Close()

	RunWithTesting(t, &RunWithTestingParams{
		Server:   srv,
		TestsDir: filepath.Join("testdata", "previous-result"),
	})
}

func testServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []byte(body)

		type nestedInfo struct {
			NestedField1 string `json:"nested_field_1"`
			NestedField2 string `json:"nested_field_2"`
		}

		if r.URL.Path == "/some/path/json" {
			w.Header().Set("Content-Type", "application/json")

			resp, _ = json.Marshal(&struct {
				Status      string     `json:"status"`
				Other       string     `json:"other"`
				UnusedField string     `json:"unused_field"`
				NestedInfo  nestedInfo `json:"nested_info"`
			}{
				Status:      "status_val",
				Other:       "some_info",
				UnusedField: "useless_info",
				NestedInfo: nestedInfo{
					NestedField1: "nested_val1",
					NestedField2: "nested_val2",
				},
			})
		}

		_, _ = w.Write(resp)
	}))
}
