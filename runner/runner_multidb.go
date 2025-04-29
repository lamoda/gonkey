package runner

import (
	"database/sql"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/joho/godotenv"

	"github.com/lamoda/gonkey/checker"
	"github.com/lamoda/gonkey/checker/response_body"
	"github.com/lamoda/gonkey/checker/response_db"
	"github.com/lamoda/gonkey/checker/response_header"
	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/fixtures/multidb"
	"github.com/lamoda/gonkey/mocks"
	"github.com/lamoda/gonkey/output"
	"github.com/lamoda/gonkey/output/allure_report"
	testingOutput "github.com/lamoda/gonkey/output/testing"
	"github.com/lamoda/gonkey/testloader/yaml_file"
	"github.com/lamoda/gonkey/variables"
)

type DbMapInstance struct {
	DB     *sql.DB
	DbType fixtures.DbType
}

type RunWithMultiDbParams struct {
	Server        *httptest.Server
	TestsDir      string
	Mocks         *mocks.Mocks
	FixturesDir   string
	DbMap         map[string]*DbMapInstance
	Aerospike     Aerospike
	EnvFilePath   string
	OutputFunc    output.OutputInterface
	Checkers      []checker.CheckerInterface
	FixtureLoader fixtures.LoaderMultiDb
}

// RunWithMultiDb is a helper function the wraps the common Run and provides simple way
// to configure Gonkey by filling the params structure.
func RunWithMultiDb(t *testing.T, params *RunWithMultiDbParams) {
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

	// Configure fixture loader
	var fixturesLoader fixtures.LoaderMultiDb

	if params.FixtureLoader == nil {
		debug := os.Getenv("GONKEY_DEBUG") != ""
		loaders := make(map[string]fixtures.Loader, len(params.DbMap))
		for connName, dbInstance := range params.DbMap {
			loaders[connName] = fixtures.NewLoader(&fixtures.Config{
				DB:       dbInstance.DB,
				Location: params.FixturesDir,
				DbType:   dbInstance.DbType,
				Debug:    debug,
			})
		}
		fixturesLoader = multidb.New(loaders)
	} else {
		fixturesLoader = params.FixtureLoader
	}

	var proxyURL *url.URL
	if os.Getenv("HTTP_PROXY") != "" {
		httpURL, err := url.Parse(os.Getenv("HTTP_PROXY"))
		if err != nil {
			t.Fatal(err)
		}
		proxyURL = httpURL
	}

	yamlLoader := yaml_file.NewLoader(params.TestsDir)
	yamlLoader.SetFileFilter(os.Getenv("GONKEY_FILE_FILTER"))

	handler := testingHandler{t}
	runner := New(
		&Config{
			Host:                  params.Server.URL,
			Mocks:                 params.Mocks,
			MocksLoader:           mocksLoader,
			FixturesLoaderMultiDb: fixturesLoader,
			Variables:             variables.New(),
			HTTPProxyURL:          proxyURL,
		},
		yamlLoader,
		handler.HandleTest,
	)

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

	runner.AddCheckers(response_body.NewChecker())
	runner.AddCheckers(response_header.NewChecker())
	runner.AddCheckers(response_db.NewMultiDbChecker(getDbConnMap(params.DbMap)))
	runner.AddCheckers(params.Checkers...)

	err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func getDbConnMap(dbMap map[string]*DbMapInstance) map[string]*sql.DB {
	result := make(map[string]*sql.DB, len(dbMap))

	for dbName, conn := range dbMap {
		result[dbName] = conn.DB
	}

	return result
}
