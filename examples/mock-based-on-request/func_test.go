package main

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/runner"
)

func TestProxy(t *testing.T) {
	m := mocks.NewNop("backend")
	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	defer m.Shutdown()

	os.Setenv("BACKEND_ADDR", m.Service("backend").ServerAddr())
	initServer()
	srv := httptest.NewServer(nil)

	runner.RunWithTesting(t, srv, &runner.RunWithTestingOpts{
		TestsDir: "cases",
		Mocks:    m,
	})
}
