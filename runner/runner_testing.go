package runner

import (
	"database/sql"
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
	"github.com/lamoda/gonkey/output/allure_report"
	testingOutput "github.com/lamoda/gonkey/output/testing"
	"github.com/lamoda/gonkey/testloader/yaml_file"
	"github.com/lamoda/gonkey/variables"
)

type RunWithTestingParams struct {
	Server      *httptest.Server
	TestsDir    string
	Mocks       *mocks.Mocks
	FixturesDir string
	DB          *sql.DB
	// If DB parameter present, used to recognize type of database, if not set, by default uses Postgres
	DbType        fixtures.DbType
	EnvFilePath   string
	OutputFunc    output.OutputInterface
	Checkers      []checker.CheckerInterface
	FixtureLoader fixtures.Loader
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
func RunWithTesting(t *testing.T, params *RunWithTestingParams) {
	var mocksLoader *mocks.Loader
	if params.Mocks != nil {
		mocksLoader = mocks.NewLoader(params.Mocks)
		registerMocksEnvironment(params.Mocks)
	}

	if params.EnvFilePath != "" {
		if err := godotenv.Load(params.EnvFilePath); err != nil {
			t.Fatal(err)
		}
	}

	debug := os.Getenv("GONKEY_DEBUG") != ""

	var fixturesLoader fixtures.Loader
	if params.DB != nil || params.FixtureLoader != nil {
		fixturesLoader = fixtures.NewLoader(&fixtures.Config{
			Location:      params.FixturesDir,
			DB:            params.DB,
			Debug:         debug,
			DbType:        params.DbType,
			FixtureLoader: params.FixtureLoader,
		})
	}

	var proxyURL *url.URL
	if os.Getenv("HTTP_PROXY") != "" {
		httpURL, err := url.Parse(os.Getenv("HTTP_PROXY"))
		if err != nil {
			t.Fatal(err)
		}
		proxyURL = httpURL
	}

	runner := initRunner(t, params, mocksLoader, fixturesLoader, proxyURL)

	if params.OutputFunc != nil {
		runner.AddOutput(params.OutputFunc)
	} else {
		runner.AddOutput(testingOutput.NewOutput())
	}

	if os.Getenv("GONKEY_ALLURE_DIR") != "" {
		allureOutput := allure_report.NewOutput("Gonkey", os.Getenv("GONKEY_ALLURE_DIR"))
		defer allureOutput.Finalize()
		runner.AddOutput(allureOutput)
	}

	addCheckers(runner, params)

	err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func initRunner(
	t *testing.T,
	params *RunWithTestingParams,
	mocksLoader *mocks.Loader,
	fixturesLoader fixtures.Loader,
	proxyURL *url.URL,
) *Runner {
	yamlLoader := yaml_file.NewLoader(params.TestsDir)
	yamlLoader.SetFileFilter(os.Getenv("GONKEY_FILE_FILTER"))

	handler := testingHandler{t}
	runner := New(
		&Config{
			Host:           params.Server.URL,
			Mocks:          params.Mocks,
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

func addCheckers(runner *Runner, params *RunWithTestingParams) {
	runner.AddCheckers(response_body.NewChecker())
	runner.AddCheckers(response_header.NewChecker())

	if params.DB != nil {
		runner.AddCheckers(response_db.NewChecker(params.DB))
	}

	runner.AddCheckers(params.Checkers...)
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
