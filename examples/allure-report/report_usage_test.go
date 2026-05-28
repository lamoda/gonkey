package allure_report

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lamoda/gonkey/runner"
)

// Example of using gonkey as a library with Allure 2 reports

func TestAPIWithAllure2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 123, "username": "newuser", "email": "user@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	os.Setenv("GONKEY_ALLURE_DIR", "./allure-results")
	os.Setenv("GONKEY_ALLURE_FORMAT", "v2")

	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:          srv,
		TestsDir:        "./simple_api_test.yaml",
		AllurePackage:   "api",
		AllureTestClass: "UsersHandler",
	})
}

func TestAPIWithAllure1Legacy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 123, "username": "newuser", "email": "user@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	os.Setenv("GONKEY_ALLURE_DIR", "./allure-results-v1")
	os.Setenv("GONKEY_ALLURE_FORMAT", "v1")

	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:   srv,
		TestsDir: "./simple_api_test.yaml",
	})
}

func TestAPIWithAllure2Default(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 123, "username": "newuser", "email": "user@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	os.Setenv("GONKEY_ALLURE_DIR", "./allure-results-default")
	os.Unsetenv("GONKEY_ALLURE_FORMAT")

	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:   srv,
		TestsDir: "./simple_api_test.yaml",
	})
}

func TestAPIWithoutAllure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/users" && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 123, "username": "newuser", "email": "user@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	os.Unsetenv("GONKEY_ALLURE_DIR")
	os.Unsetenv("GONKEY_ALLURE_FORMAT")

	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:   srv,
		TestsDir: "./simple_api_test.yaml",
	})
}

func TestAPIWithAllure2_TestITLabels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/users/123" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": 123, "username": "testuser", "email": "test@example.com"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	os.Setenv("GONKEY_ALLURE_DIR", "./allure-results-without-labels")
	os.Setenv("GONKEY_ALLURE_FORMAT", "v2")

	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:          srv,
		TestsDir:        "without_labels_test.yaml",
		AllurePackage:   "api",
		AllureTestClass: "UsersHandler",
	})
}
