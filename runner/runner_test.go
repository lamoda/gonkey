package runner

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/lamoda/gonkey/models"
	"github.com/stretchr/testify/assert"
)

func TestDontFollowRedirects(t *testing.T) {
	srv := testServerRedirect()
	defer srv.Close()

	RunWithTesting(t, &RunWithTestingParams{
		Server:   srv,
		TestsDir: filepath.Join("testdata", "dont-follow-redirects"),
	})
}

func testServerRedirect() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redirect-url", http.StatusFound)
	}))
}

func TestBuildAllureDefaultLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params *RunWithTestingParams
		want   []models.AllureLabel
	}{
		{
			name: "legacy fields override custom labels",
			params: &RunWithTestingParams{
				AllureDefaultLabels: []models.AllureLabel{
					{Name: "layer", Value: "Integration"},
					{Name: "package", Value: "cron"},
				},
				AllurePackage:   "api",
				AllureTestClass: "UsersHandler",
			},
			want: []models.AllureLabel{
				{Name: "layer", Value: "Integration"},
				{Name: "package", Value: "api"},
				{Name: "testClass", Value: "UsersHandler"},
			},
		},
		{
			name: "empty legacy fields do not overwrite custom package/testClass",
			params: &RunWithTestingParams{
				AllureDefaultLabels: []models.AllureLabel{
					{Name: "layer", Value: "Integration"},
					{Name: "package", Value: "cron"},
					{Name: "testClass", Value: "CronHandler"},
				},
			},
			want: []models.AllureLabel{
				{Name: "layer", Value: "Integration"},
				{Name: "package", Value: "cron"},
				{Name: "testClass", Value: "CronHandler"},
			},
		},
		{
			name: "empty name or value labels are dropped",
			params: &RunWithTestingParams{
				AllureDefaultLabels: []models.AllureLabel{
					{Name: "layer", Value: "Integration"},
					{Name: "team", Value: ""},
					{Name: "", Value: "Orders shipment"},
				},
				AllurePackage:   "api",
				AllureTestClass: "UsersHandler",
			},
			want: []models.AllureLabel{
				{Name: "layer", Value: "Integration"},
				{Name: "package", Value: "api"},
				{Name: "testClass", Value: "UsersHandler"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, buildAllureDefaultLabels(tt.params))
		})
	}
}
