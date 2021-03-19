package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/lamoda/gonkey/checker/response_body"
	"github.com/lamoda/gonkey/checker/response_db"
	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/output/allure_report"
	"github.com/lamoda/gonkey/output/console_colored"
	"github.com/lamoda/gonkey/runner"
	"github.com/lamoda/gonkey/testloader/yaml_file"
	"github.com/lamoda/gonkey/variables"
)

func main() {
	var config struct {
		Host             string
		TestsLocation    string
		DbDsn            string
		FixturesLocation string
		EnvFile          string
		Allure           bool
		Verbose          bool
		Debug            bool
		DbType           string
	}

	flag.StringVar(&config.Host, "host", "", "Target system hostname")
	flag.StringVar(&config.TestsLocation, "tests", "", "Path to tests file or directory")
	flag.StringVar(&config.DbDsn, "db_dsn", "", "DSN for the fixtures database (WARNING! Db tables will be truncated)")
	flag.StringVar(&config.FixturesLocation, "fixtures", "", "Path to fixtures directory")
	flag.StringVar(&config.EnvFile, "env-file", "", "Path to env-file")
	flag.BoolVar(&config.Allure, "allure", true, "Make Allure report")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&config.Debug, "debug", false, "Debug output")
	flag.StringVar(
		&config.EnvFile,
		"db-type",
		fixtures.PostgresParam,
		"Type of database (options: postgres, mysql)",
	)

	flag.Parse()

	if config.Host == "" {
		log.Fatal(errors.New("service hostname not provided"))
	} else {
		if !strings.HasPrefix(config.Host, "http://") && !strings.HasPrefix(config.Host, "https://") {
			config.Host = "http://" + config.Host
		}
		config.Host = strings.TrimRight(config.Host, "/")
	}

	if config.TestsLocation == "" {
		log.Fatal(errors.New("no tests location provided"))
	}

	var db *sql.DB
	if config.DbDsn != "" {
		var err error
		db, err = sql.Open("postgres", config.DbDsn)
		if err != nil {
			log.Fatal(err)
		}
	}

	var fixturesLoader fixtures.Loader
	if db != nil && config.FixturesLocation != "" {
		fixturesLoader = fixtures.NewLoader(&fixtures.Config{
			DB:       db,
			Location: config.FixturesLocation,
			Debug:    config.Debug,
			DbType:   fixtures.FetchDbType(config.DbType),
		})
	} else if config.FixturesLocation != "" {
		log.Fatal(errors.New("you should specify db_dsn to load fixtures"))
	}

	if config.EnvFile != "" {
		if err := godotenv.Load(config.EnvFile); err != nil {
			log.Println(errors.New("can't load .env file"), err)
		}
	}

	r := runner.New(
		&runner.Config{
			Host:           config.Host,
			FixturesLoader: fixturesLoader,
			Variables:      variables.New(),
		},
		yaml_file.NewLoader(config.TestsLocation),
	)

	consoleOutput := console_colored.NewOutput(config.Verbose)
	r.AddOutput(consoleOutput)

	var allureOutput *allure_report.AllureReportOutput
	if config.Allure {
		allureOutput = allure_report.NewOutput("Gonkey", "./allure-results")
		r.AddOutput(allureOutput)
	}

	r.AddCheckers(response_body.NewChecker())

	if db != nil {
		r.AddCheckers(response_db.NewChecker(db))
	}

	summary, err := r.Run()
	if err != nil {
		log.Fatal(err)
	}

	consoleOutput.ShowSummary(summary)

	if allureOutput != nil {
		allureOutput.Finalize()
	}

	if !summary.Success {
		os.Exit(1)
	}
}
