package storage

import (
	"encoding/json"
)

type StorageInterface interface {
	Type() string
	LoadFixtures(location string, names []string) error
	ExecuteQuery(query string) ([]json.RawMessage, error)
}
