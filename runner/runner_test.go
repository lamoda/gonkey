package runner

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestDontFollowRedirects(t *testing.T) {
	srv := testServerRedirect()
	defer srv.Close()

	RunWithTesting(t, srv, &RunWithTestingOpts{
		TestsDir: filepath.Join("testdata", "dont-follow-redirects"),
	})
}

func testServerRedirect() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect-url", http.StatusFound)
	}))
}
