package fixtures

import (
	"fmt"

	"github.com/lamoda/gonkey/storage"
)

type Loader interface {
	Load(names []string) error
}

func NewLoader(location string, storages []storage.StorageInterface) Loader {
	return &loaderImpl{
		location: location,
		storages: storages,
	}
}

type loaderImpl struct {
	location string
	storages []storage.StorageInterface
}

func (l *loaderImpl) Load(names []string) error {
	for _, s := range l.storages {
		err := s.LoadFixtures(l.location, names)
		if err != nil {
			return fmt.Errorf("load fixtures: %w", err)
		}
	}
	return nil
}
