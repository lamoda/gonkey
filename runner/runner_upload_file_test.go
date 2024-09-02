package runner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadFiles(t *testing.T) {
	srv := testServerUpload(t)
	defer srv.Close()

	// TODO: refactor RunWithTesting() for testing negative scenario (when tests has expected errors)
	RunWithTesting(t, &RunWithTestingParams{
		Server:   srv,
		TestsDir: filepath.Join("testdata", "upload-files"),
	})
}

type response struct {
	Status       string `json:"status"`
	File1Name    string `json:"file_1_name"`
	File1Content string `json:"file_1_content"`
	File2Name    string `json:"file_2_name"`
	File2Content string `json:"file_2_content"`
}

func testServerUpload(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := response{
			Status: "OK",
		}

		resp.File1Name, resp.File1Content = formFile(t, r, "file1")
		resp.File2Name, resp.File2Content = formFile(t, r, "file2")

		respData, err := json.Marshal(resp)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")

		_, err = w.Write(respData)
		require.NoError(t, err)
	}))
}

func formFile(t *testing.T, r *http.Request, field string) (string, string) {
	file, header, err := r.FormFile(field)
	require.NoError(t, err)

	defer func() { _ = file.Close() }()

	contents, err := ioutil.ReadAll(file)
	require.NoError(t, err)

	return header.Filename, string(contents)
}
