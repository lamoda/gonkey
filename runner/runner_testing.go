package runner

import (
	"database/sql"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aerospike/aerospike-client-go/v5"
	"github.com/joho/godotenv"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/checker/response_body"
	"github.com/lamoda/gonkey/checker/response_db"
	"github.com/lamoda/gonkey/checker/response_header"
	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/output"
	"github.com/lamoda/gonkey/output/allure_report"
	testingOutput "github.com/lamoda/gonkey/output/testing"
	aerospikeAdapter "github.com/lamoda/gonkey/storage/aerospike"
	"github.com/lamoda/gonkey/testloader/yaml_file"
	"github.com/lamoda/gonkey/variables"
)

type Aerospike struct {
	*aerospike.Client
	Namespace string
}

type RunWithTestingParams struct {
	Server        *httptest.Server
	TestsDir      string
	Mocks         *mocks.Mocks
	FixturesDir   string
	DB            *sql.DB
	Aerospike     Aerospike
	// If DB parameter present, used to recognize type of database, if not set, by default uses Postgres
	DbType        fixtures.DbType
	EnvFilePath   string
	OutputFunc    output.OutputInterface
	Checkers      []checker.CheckerInterface
	FixtureLoader fixtures.Loader
}

// RunWithTesting is a helper function the wraps the common Run and provides simple way
// to configure Gonkey by filling the params structure.
func RunWithTesting(t *testing.T, params *RunWithTestingParams) {
	var mocksLoader *mocks.Loader
	if params.Mocks != nil {
		mocksLoader = mocks.NewLoader(params.Mocks)
	}

	if params.EnvFilePath != "" {
		if err := godotenv.Load(params.EnvFilePath); err != nil {
			t.Fatal(err)
		}
	}

	debug := os.Getenv("GONKEY_DEBUG") != ""

	var fixturesLoader fixtures.Loader
	if params.DB != nil || params.Aerospike.Client != nil || params.FixtureLoader != nil  {
		fixturesLoader = fixtures.NewLoader(&fixtures.Config{
			Location:      params.FixturesDir,
			DB:            params.DB,
			Aerospike:     aerospikeAdapter.New(params.Aerospike.Client, params.Aerospike.Namespace),
			Debug:         debug,
			DbType:        params.DbType,
			FixtureLoader: params.FixtureLoader,
		})
	}

	runner := initRunner(params, mocksLoader, fixturesLoader)

	setupOutputs(runner, params, t)

	addCheckers(runner, params)

	_, err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func initRunner(params *RunWithTestingParams, mocksLoader *mocks.Loader, fixturesLoader fixtures.Loader) *Runner {
	yamlLoader := yaml_file.NewLoader(params.TestsDir)
	yamlLoader.SetFileFilter(os.Getenv("GONKEY_FILE_FILTER"))

	runner := New(
		&Config{
			Host:           params.Server.URL,
			Mocks:          params.Mocks,
			MocksLoader:    mocksLoader,
			FixturesLoader: fixturesLoader,
			Variables:      variables.New(),
		},
		yamlLoader,
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

func setupOutputs(r *Runner, params *RunWithTestingParams, t *testing.T) {
	if params.OutputFunc != nil {
		r.AddOutput(params.OutputFunc)
	} else {
		r.AddOutput(testingOutput.NewOutput(t))
	}

	if os.Getenv("GONKEY_ALLURE_DIR") != "" {
		allureOutput := allure_report.NewOutput("Gonkey", os.Getenv("GONKEY_ALLURE_DIR"))
		defer allureOutput.Finalize()
		r.AddOutput(allureOutput)
	}
}
