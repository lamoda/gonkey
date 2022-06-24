package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"github.com/lamoda/gonkey/checker/response_body"
	"github.com/lamoda/gonkey/checker/response_db"
	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/output/allure_report"
	"github.com/lamoda/gonkey/output/console_colored"
	"github.com/lamoda/gonkey/runner"
	"github.com/lamoda/gonkey/storage/aerospike"
	"github.com/lamoda/gonkey/testloader/yaml_file"
	"github.com/lamoda/gonkey/variables"
)

type config struct {
	Host             string
	TestsLocation    string
	DbDsn            string
	AerospikeHost    string
	FixturesLocation string
	EnvFile          string
	Allure           bool
	Verbose          bool
	Debug            bool
	DbType           string
}

type storages struct {
	db        *sql.DB
	aerospike *aerospike.Client
}

func main() {
	cfg := getConfig()
	validateConfig(&cfg)

	storages := initStorages(cfg)

	fixturesLoader := initLoaders(storages, cfg)

	r := initRunner(cfg, fixturesLoader)

	setupOutputs(r, cfg)

	addCheckers(r, storages.db)
}

func initStorages(cfg config) storages {
	db := initDB(cfg)
	aerospikeClient := initAerospike(cfg)
	return storages{
		db:        db,
		aerospike: aerospikeClient,
	}
}

func initLoaders(storages storages, cfg config) fixtures.Loader {
	var fixturesLoader fixtures.Loader
	if (storages.db != nil || storages.aerospike != nil) && cfg.FixturesLocation != "" {
		fixturesLoader = fixtures.NewLoader(&fixtures.Config{
			DB:        storages.db,
			Aerospike: storages.aerospike,
			Location:  cfg.FixturesLocation,
			Debug:     cfg.Debug,
			DbType:    fixtures.FetchDbType(cfg.DbType),
		})
	} else if cfg.FixturesLocation != "" {
		log.Fatal(errors.New("you should specify db_dsn to load fixtures"))
	}
	return fixturesLoader
}

func validateConfig(cfg *config) {
	if cfg.Host == "" {
		log.Fatal(errors.New("service hostname not provided"))
	} else {
		if !strings.HasPrefix(cfg.Host, "http://") && !strings.HasPrefix(cfg.Host, "https://") {
			cfg.Host = "http://" + cfg.Host
		}
		cfg.Host = strings.TrimRight(cfg.Host, "/")
	}

	if cfg.TestsLocation == "" {
		log.Fatal(errors.New("no tests location provided"))
	}

	if cfg.EnvFile != "" {
		if err := godotenv.Load(cfg.EnvFile); err != nil {
			log.Println(errors.New("can't load .env file"), err)
		}
	}
}

func addCheckers(r *runner.Runner, db *sql.DB) {
	r.AddCheckers(response_body.NewChecker())
	if db != nil {
		r.AddCheckers(response_db.NewChecker(db))
	}
}

func setupOutputs(r *runner.Runner, cfg config) {
	consoleOutput := console_colored.NewOutput(cfg.Verbose)
	r.AddOutput(consoleOutput)

	var allureOutput *allure_report.AllureReportOutput
	if cfg.Allure {
		allureOutput = allure_report.NewOutput("Gonkey", "./allure-results")
		r.AddOutput(allureOutput)
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

func initRunner(cfg config, fixturesLoader fixtures.Loader) *runner.Runner {
	return runner.New(
		&runner.Config{
			Host:           cfg.Host,
			FixturesLoader: fixturesLoader,
			Variables:      variables.New(),
		},
		yaml_file.NewLoader(cfg.TestsLocation),
	)
}

func initAerospike(cfg config) *aerospike.Client {
	if cfg.AerospikeHost != "" {
		address, port, namespace := parseAerospikeHost(cfg.AerospikeHost)
		return aerospike.New(address, port, namespace)
	}

	return nil
}

func initDB(cfg config) *sql.DB {
	if cfg.DbDsn != "" {
		var err error
		db, err := sql.Open("postgres", cfg.DbDsn)
		if err != nil {
			log.Fatal(err)
		}
		return db
	}

	return nil
}

func getConfig() config {
	cfg := config{}

	flag.StringVar(&cfg.Host, "host", "", "Target system hostname")
	flag.StringVar(&cfg.TestsLocation, "tests", "", "Path to tests file or directory")
	flag.StringVar(&cfg.DbDsn, "db_dsn", "", "DSN for the fixtures database (WARNING! Db tables will be truncated)")
	flag.StringVar(&cfg.AerospikeHost, "aerospike_host", "", "Aerospike host for fixtures in form of 'host:port/namespace' (WARNING! Aerospike sets will be truncated)")
	flag.StringVar(&cfg.FixturesLocation, "fixtures", "", "Path to fixtures directory")
	flag.StringVar(&cfg.EnvFile, "env-file", "", "Path to env-file")
	flag.BoolVar(&cfg.Allure, "allure", true, "Make Allure report")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.Debug, "debug", false, "Debug output")
	flag.StringVar(
		&cfg.DbType,
		"db-type",
		fixtures.PostgresParam,
		"Type of database (options: postgres, mysql, aerospike)",
	)

	flag.Parse()
	return cfg
}

func parseAerospikeHost(dsn string) (address string, port int, namespace string) {
	parts := strings.Split(dsn, "/")
	if len(parts) != 2 {
		log.Fatalf("couldn't parse aerospike host %v, should be in form of host:port/namespace", dsn)
	}
	namespace = parts[1]

	host := parts[0]
	hostParts := strings.Split(host, ":")
	address = hostParts[0]
	port, err := strconv.Atoi(hostParts[1])
	if err != nil {
		log.Fatal("couldn't parse port: " + parts[1])
	}

	return
}
