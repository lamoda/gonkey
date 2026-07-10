package runner

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lamoda/gonkey/mocks"
)

func Test_MockVariablesToSet(t *testing.T) {
	m := mocks.NewNop("backend")
	require.NoError(t, m.Start())
	defer m.Shutdown()

	callMock := func(path, body string, headers map[string]string) {
		mockURL := "http://" + os.Getenv("GONKEY_MOCK_BACKEND") + path
		req, err := http.NewRequest(http.MethodPost, mockURL, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/register":
			var req struct {
				UserID string `json:"user_id"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

			callMock("/notify?source=gonkey", `{"user": {"id": "`+req.UserID+`"}}`,
				map[string]string{"X-Request-Id": "req-100500"})

			w.Write([]byte(`{"ok": true}`))
		case "/register-collision":
			var req struct {
				UserID string `json:"user_id"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

			callMock("/notify", `{"user": {"id": "`+req.UserID+`"}}`, nil)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id": "from-response"}`))
		case "/transfer":
			var req struct {
				From string `json:"from"`
				To   string `json:"to"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

			callMock("/debit", `{"account": "`+req.From+`"}`, nil)
			callMock("/credit", `{"account": "`+req.To+`"}`, nil)

			w.Write([]byte(`{"ok": true}`))
		case "/echo":
			io.Copy(w, r.Body)
		}
	}))
	defer srv.Close()

	RunWithTesting(t, &RunWithTestingParams{
		Server:   srv,
		TestsDir: filepath.Join("testdata", "mock-variables"),
		Mocks:    m,
	})
}
