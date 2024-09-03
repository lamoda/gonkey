package runner

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/checker/response_body"
	"github.com/lamoda/gonkey/checker/response_db"
	"github.com/lamoda/gonkey/checker/response_header"
	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/models"
	"github.com/lamoda/gonkey/output"
	testingOutput "github.com/lamoda/gonkey/output/testing"
	"github.com/lamoda/gonkey/storage"
	"github.com/lamoda/gonkey/testloader/yaml_file"
	"github.com/lamoda/gonkey/variables"
)

var DefaultOutput = testingOutput.NewOutput()

type RunWithTestingOpts struct {
	TestsDir    string
	FixturesDir string
	EnvFilePath string

	Mocks          *mocks.Mocks
	Storages       []storage.StorageInterface
	MainOutputFunc output.OutputInterface
	Outputs        []output.OutputInterface
	Checkers       []checker.CheckerInterface
}

func registerMocksEnvironment(m *mocks.Mocks) {
	names := m.GetNames()
	for _, n := range names {
		varName := fmt.Sprintf("GONKEY_MOCK_%s", strings.ToUpper(n))
		os.Setenv(varName, m.Service(n).ServerAddr())
	}
}

// RunWithTesting is a helper function the wraps the common Run and provides simple way
// to configure Gonkey by filling the params structure.
func RunWithTesting(t *testing.T, server *httptest.Server, opts *RunWithTestingOpts) {
	var mocksLoader *mocks.Loader
	if opts.Mocks != nil {
		mocksLoader = mocks.NewLoader(opts.Mocks)
		registerMocksEnvironment(opts.Mocks)
	}

	if opts.EnvFilePath != "" {
		if err := godotenv.Load(opts.EnvFilePath); err != nil {
			t.Fatal(err)
		}
	}

	fixturesLoader := fixtures.NewLoader(opts.FixturesDir, opts.Storages)

	var proxyURL *url.URL
	if os.Getenv("HTTP_PROXY") != "" {
		httpURL, err := url.Parse(os.Getenv("HTTP_PROXY"))
		if err != nil {
			t.Fatal(err)
		}
		proxyURL = httpURL
	}

	runner := initRunner(t, server, opts, mocksLoader, fixturesLoader, proxyURL)

	addOutputs(runner, opts)
	addCheckers(runner, opts)

	err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func initRunner(
	t *testing.T,
	server *httptest.Server,
	opts *RunWithTestingOpts,
	mocksLoader *mocks.Loader,
	fixturesLoader fixtures.Loader,
	proxyURL *url.URL,
) *Runner {
	yamlLoader := yaml_file.NewLoader(opts.TestsDir)
	yamlLoader.SetFileFilter(os.Getenv("GONKEY_FILE_FILTER"))

	handler := testingHandler{t}
	runner := New(
		&Config{
			Host:           server.URL,
			Mocks:          opts.Mocks,
			MocksLoader:    mocksLoader,
			FixturesLoader: fixturesLoader,
			Variables:      variables.New(),
			HTTPProxyURL:   proxyURL,
		},
		yamlLoader,
		handler.HandleTest,
	)

	return runner
}

func addOutputs(runner *Runner, opts *RunWithTestingOpts) {
	if opts.MainOutputFunc != nil {
		runner.AddOutput(opts.MainOutputFunc)
	} else {
		runner.AddOutput(DefaultOutput)
	}

	for _, o := range opts.Outputs {
		runner.AddOutput(o)
	}
}

func addCheckers(runner *Runner, opts *RunWithTestingOpts) {
	runner.AddCheckers(response_body.NewChecker())
	runner.AddCheckers(response_header.NewChecker())
	if len(opts.Storages) != 0 {
		runner.AddCheckers(response_db.NewChecker(opts.Storages))
	}
	runner.AddCheckers(opts.Checkers...)
}

type testingHandler struct {
	t *testing.T
}

func (h testingHandler) HandleTest(test models.TestInterface, executeTest testExecutor) error {
	var returnErr error
	h.t.Run(test.GetName(), func(t *testing.T) {
		result, err := executeTest(test)
		if err != nil {
			if errors.Is(err, errTestSkipped) || errors.Is(err, errTestBroken) {
				t.Skip()
			} else {
				returnErr = err
				t.Fatal(err)
			}
		}

		if !result.Passed() {
			t.Fail()
		}
	})

	return returnErr
}
