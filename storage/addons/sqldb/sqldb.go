package sqldb

import (
	"database/sql"
	"encoding/json"

	"github.com/lamoda/gonkey/storage/addons/sqldb/mysql"
	"github.com/lamoda/gonkey/storage/addons/sqldb/postgresql"
)

type SQLType string

const (
	PostgreSQL SQLType = "postgresql"
	MySQL      SQLType = "mysql"
)

type Storage struct {
	dbType SQLType
	db     *sql.DB
}

type StorageOpts struct {
}

func NewStorage(dbType SQLType, db *sql.DB, opts StorageOpts) (*Storage, error) {
	if dbType != PostgreSQL && dbType != MySQL {
		panic("unknown db type param")
	}
	return &Storage{
		dbType: dbType,
		db:     db,
	}, nil
}

func (l *Storage) Type() string {
	return string(l.dbType)
}

func (l *Storage) LoadFixtures(location string, names []string) error {
	if l.dbType == PostgreSQL {
		return postgresql.LoadFixtures(l.db, location, names)
	}
	return mysql.LoadFixtures(l.db, location, names)
}

func (l *Storage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	if l.dbType == PostgreSQL {
		return postgresql.ExecuteQuery(l.db, query)
	}
	return mysql.ExecuteQuery(l.db, query)
}
