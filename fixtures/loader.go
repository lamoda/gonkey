package fixtures

import (
	"database/sql"
	"strings"

	aerospikeClient "github.com/aerospike/aerospike-client-go/v5"
	_ "github.com/lib/pq"

	"github.com/lamoda/gonkey/fixtures/aerospike"
	"github.com/lamoda/gonkey/fixtures/mysql"
	"github.com/lamoda/gonkey/fixtures/postgres"
)

type DbType int

const (
	Postgres DbType = iota
	Mysql
	Aerospike
)

const (
	PostgresParam  = "postgres"
	MysqlParam     = "mysql"
	AerospikeParam = "aerospike"
)

type Config struct {
	DB        *sql.DB
	Aerospike *aerospikeClient.Client
	DbType    DbType
	Location  string
	Debug     bool
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
	default:
		panic("unknown db type param")
	}
}
