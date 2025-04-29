package multidb

import (
	"fmt"

	"github.com/lamoda/gonkey/fixtures"
	"github.com/lamoda/gonkey/models"
)

type LoaderByMap struct {
	loaders map[string]fixtures.Loader
}

func New(loaders map[string]fixtures.Loader) *LoaderByMap {
	return &LoaderByMap{
		loaders: loaders,
	}
}

func (l *LoaderByMap) Load(fixturesList models.FixturesMultiDb) error {
	for _, fixture := range fixturesList {
		loader, ok := l.loaders[fixture.DbName]
		if !ok {
			return fmt.Errorf("loader %s not exists", fixture.DbName)
		}

		if err := loader.Load(fixture.Files); err != nil {
			return err
		}
	}

	return nil
}
