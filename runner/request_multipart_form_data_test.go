package runner

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMultipartFormData(t *testing.T) {
	srv := testServerMultipartFormData(t)
	defer srv.Close()

	// TODO: refactor RunWithTesting() for testing negative scenario (when tests has expected errors)
	RunWithTesting(t, &RunWithTestingParams{
		Server:   srv,
		TestsDir: filepath.Join("testdata", "multipart", "form-data"),
	})
}

type multipartResponse struct {
	ContentTypeHeader  string `json:"content_type_header"`
	RequestBodyContent string `json:"request_body_content"`
}

func testServerMultipartFormData(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		r.Body = io.NopCloser(bytes.NewReader(body))

		resp := multipartResponse{
			ContentTypeHeader:  r.Header.Get("Content-Type"),
			RequestBodyContent: string(body),
		}

		respData, err := json.Marshal(resp)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(respData)
		require.NoError(t, err)

	}))
}
