package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/lamoda/gonkey/checker/response_body"
	"github.com/lamoda/gonkey/checker/response_schema"
	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/output/allure_report"
	"github.com/lamoda/gonkey/output/console_colored"
	"github.com/lamoda/gonkey/runner"
	"github.com/lamoda/gonkey/testloader/yaml_file"
)

func main() {
	var config struct {
		Host             string
		SpecPath         string
		TestsLocation    string
		DbDsn            string
		FixturesLocation string
		Allure           bool
		Verbose          bool
		Debug            bool
	}

	flag.StringVar(&config.Host, "host", "", "Target system hostname")
	flag.StringVar(&config.SpecPath, "spec", "", "Path or URL to swagger specification")
	flag.StringVar(&config.TestsLocation, "tests", "", "Path to tests file or directory")
	flag.StringVar(&config.DbDsn, "db_dsn", "", "DSN for the fixtures database (WARNING! Db tables will be truncated)")
	flag.StringVar(&config.FixturesLocation, "fixtures", "", "Path to fixtures directory")
	flag.BoolVar(&config.Allure, "allure", true, "Make Allure report")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&config.Debug, "debug", false, "Debug output")

	flag.Parse()

	if config.Host == "" {
		log.Fatal(errors.New("service hostname not provided"))
	} else {
		if !strings.HasPrefix(config.Host, "http://") {
			config.Host = "http://" + config.Host
		}
		config.Host = strings.TrimRight(config.Host, "/")
	}

	if config.TestsLocation == "" {
		log.Fatal(errors.New("no tests location provided"))
	}

	var fixturesLoader *fixtures.Loader
	if config.DbDsn != "" && config.FixturesLocation != "" {
		db, err := sql.Open("postgres", config.DbDsn)
		if err != nil {
			log.Fatal(err)
		}
		fixturesLoader = fixtures.NewLoader(&fixtures.Config{
			DB:       db,
			Location: config.FixturesLocation,
			Debug:    config.Debug,
		})
	} else if !(config.DbDsn == "" && config.FixturesLocation == "") {
		log.Fatal(errors.New("you should specify either both db_dsn and fixtures or none of them"))
	}

	r := runner.New(
		&runner.Config{
			Host:           config.Host,
			FixturesLoader: fixturesLoader,
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
	if config.SpecPath != "" {
		r.AddCheckers(response_schema.NewChecker(config.SpecPath))
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
