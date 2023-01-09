package fixtures

import (
	"database/sql"
	"strings"

	_ "github.com/lib/pq"

	"github.com/lamoda/gonkey/fixtures/aerospike"
	"github.com/lamoda/gonkey/fixtures/mysql"
	"github.com/lamoda/gonkey/fixtures/postgres"
	aerospikeClient "github.com/lamoda/gonkey/storage/aerospike"
)

type DbType int

const (
	Postgres DbType = iota
	Mysql
	Aerospike
	Redis
	CustomLoader // using external loader if gonkey used as a library
)

const (
	PostgresParam  = "postgres"
	MysqlParam     = "mysql"
	AerospikeParam = "aerospike"
	RedisParam     = "redis"
)

type Config struct {
	DB            *sql.DB
	Aerospike     *aerospikeClient.Client
	DbType        DbType
	Location      string
	Debug         bool
	FixtureLoader Loader
}

type Loader interface {
	Load(names []string) error
}

func NewLoader(cfg *Config) Loader {

	var loader Loader

	location := strings.TrimRight(cfg.Location, "/")

	switch cfg.DbType {
	case Postgres:
		loader = postgres.New(
			cfg.DB,
			location,
			cfg.Debug,
		)
	case Mysql:
		loader = mysql.New(
			cfg.DB,
			location,
			cfg.Debug,
		)
	case Aerospike:
		loader = aerospike.New(
			cfg.Aerospike,
			location,
			cfg.Debug,
		)
	default:
		if cfg.FixtureLoader != nil {
			return cfg.FixtureLoader
		}
		panic("unknown db type")
	}

	return loader
}

func FetchDbType(dbType string) DbType {
	switch dbType {
	case PostgresParam:
		return Postgres
	case MysqlParam:
		return Mysql
	case AerospikeParam:
		return Aerospike
	case RedisParam:
		return Redis
	default:
		panic("unknown db type param")
	}
}
