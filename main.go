package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/aerospike/aerospike-client-go/v5"
	"github.com/go-redis/redis/v9"
	"github.com/joho/godotenv"

	"github.com/lamoda/gonkey/checker/response_body"
	"github.com/lamoda/gonkey/checker/response_db"
	"github.com/lamoda/gonkey/fixtures"
	redisLoader "github.com/lamoda/gonkey/fixtures/redis"
	"github.com/lamoda/gonkey/output/allure_report"
	"github.com/lamoda/gonkey/output/console_colored"
	"github.com/lamoda/gonkey/runner"
	aerospikeAdapter "github.com/lamoda/gonkey/storage/aerospike"
	mongoAdapter "github.com/lamoda/gonkey/storage/mongo"
	"github.com/lamoda/gonkey/testloader/yaml_file"
	"github.com/lamoda/gonkey/variables"
)

type config struct {
	Host             string
	TestsLocation    string
	DbDsn            string
	AerospikeHost    string
	MongoDsn         string
	RedisURL         string
	FixturesLocation string
	EnvFile          string
	Allure           bool
	Verbose          bool
	Debug            bool
	DbType           string
}

type storages struct {
	db        *sql.DB
	aerospike *aerospikeAdapter.Client
	mongo     *mongoAdapter.Client
}

func main() {
	cfg := getConfig()
	validateConfig(&cfg)

	storages := initStorages(cfg)

	testHandler := runner.NewConsoleHandler()
	fixturesLoader := initLoaders(storages, cfg)

	proxyURL, err := proxyURLFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	testsRunner := initRunner(cfg, fixturesLoader, testHandler, proxyURL)

	consoleOutput := console_colored.NewOutput(cfg.Verbose)
	testsRunner.AddOutput(consoleOutput)

	addCheckers(testsRunner, storages.db)

	var allureOutput *allure_report.AllureReportOutput
	if cfg.Allure {
		allureOutput = allure_report.NewOutput("Gonkey", "./allure-results")
		testsRunner.AddOutput(allureOutput)
	}

	err = testsRunner.Run()
	if err != nil {
		log.Fatal(err)
	}

	if allureOutput != nil {
		allureOutput.Finalize()
	}

	summary := testHandler.Summary()
	consoleOutput.ShowSummary(summary)
	if !summary.Success {
		os.Exit(1)
	}
}

func initStorages(cfg config) storages {
	db := initDB(cfg)
	aerospikeClient := initAerospike(cfg)
	mongoClient := initMongo(cfg)
	return storages{
		db:        db,
		aerospike: aerospikeClient,
		mongo:     mongoClient,
	}
}

func initLoaders(storages storages, cfg config) fixtures.Loader {
	var fixturesLoader fixtures.Loader
	if cfg.FixturesLocation != "" {
		if storages.db != nil || storages.aerospike != nil || storages.mongo != nil {
			fixturesLoader = fixtures.NewLoader(&fixtures.Config{
				DB:        storages.db,
				Aerospike: storages.aerospike,
				Mongo:     storages.mongo,
				Location:  cfg.FixturesLocation,
				Debug:     cfg.Debug,
				DbType:    fixtures.FetchDbType(cfg.DbType),
			})
		} else if cfg.DbType == fixtures.RedisParam {
			redisOptions, err := redis.ParseURL(cfg.RedisURL)
			if err != nil {
				log.Panic("redis_url attribute is not a valid URL")
			}
			fixturesLoader = redisLoader.New(redisLoader.LoaderOptions{
				FixtureDir: cfg.FixturesLocation,
				Redis:      redisOptions,
			})
		} else {
			log.Fatal(errors.New("you should specify db_dsn to load fixtures"))
		}
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

func initRunner(
	cfg config,
	fixturesLoader fixtures.Loader,
	handler *runner.ConsoleHandler,
	proxyURL *url.URL,
) *runner.Runner {
	return runner.New(
		&runner.Config{
			Host:           cfg.Host,
			FixturesLoader: fixturesLoader,
			Variables:      variables.New(),
			HttpProxyURL:   proxyURL,
		},
		yaml_file.NewLoader(cfg.TestsLocation),
		handler.HandleTest,
	)
}

func initAerospike(cfg config) *aerospikeAdapter.Client {
	if cfg.AerospikeHost != "" {
		address, port, namespace := parseAerospikeHost(cfg.AerospikeHost)
		client, err := aerospike.NewClient(address, port)
		if err != nil {
			log.Fatal("Couldn't connect to aerospike: ", err)
		}
		return aerospikeAdapter.New(client, namespace)
	}

	return nil
}

func initMongo(cfg config) *mongoAdapter.Client {
	if cfg.MongoDsn != "" {
		client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoDsn))
		if err != nil {
			log.Fatal("Couldn't connect to mongo: ", err)
		}

		return mongoAdapter.New(client)
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
	flag.StringVar(&cfg.MongoDsn, "mongo_dsn", "", "Mongo DSN for the fixtures database (WARNING! Mongo collections will be truncated)")
	flag.StringVar(&cfg.RedisURL, "redis_url", "", "Redis server URL for fixture loading")
	flag.StringVar(&cfg.FixturesLocation, "fixtures", "", "Path to fixtures directory")
	flag.StringVar(&cfg.EnvFile, "env-file", "", "Path to env-file")
	flag.BoolVar(&cfg.Allure, "allure", true, "Make Allure report")
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.Debug, "debug", false, "Debug output")
	flag.StringVar(
		&cfg.DbType,
		"db-type",
		fixtures.PostgresParam,
		"Type of database (options: postgres, mysql, aerospike, redis, mongo)",
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

func proxyURLFromEnv() (*url.URL, error) {
	if os.Getenv("HTTP_PROXY") != "" {
		httpUrl, err := url.Parse(os.Getenv("HTTP_PROXY"))
		if err != nil {
			return nil, err
		}
		return httpUrl, nil
	}
	return nil, nil
}
