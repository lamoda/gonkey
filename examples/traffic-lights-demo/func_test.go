package main

import (
	"net/http/httptest"
	"testing"

	"github.com/lamoda/gonkey/runner"
)

func Test_API(t *testing.T) {
	initServer()

	srv := httptest.NewServer(nil)

	runner.RunWithTesting(t, &runner.RunWithTestingParams{
		Server:   srv,
		TestsDir: "tests/cases",
	})
}
