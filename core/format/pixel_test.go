package format

import (
	"Pixel/config"
	"Pixel/store/engine"
	"Pixel/store/service"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"testing"
)

func TestPixel_Parse(t *testing.T) {
	c := config.Config{
		Files: config.Files{
			SourceFolder:  "./data",
			WorkingFolder: "./data/ext",
		},
		Store: config.StoreGroup{
			Path:    "./data",
			Timeout: "120",
		},
	}
	pixel := Pixel(c, makeDateService(c))
	pixel.Parse()
}

func makeDateService(c config.Config) *service.DataStore {
	storeEngine, err := makeDataStore(c)
	if err != nil {
		log.Printf("[ERR] Ошибка подключения к SQLite")
	}
	return &service.DataStore{
		Engine: storeEngine,
	}
}

func makeDataStore(c config.Config) (result engine.Interface, err error) {
	if err = makeDirs(c.Store.Path); err != nil {
		return nil, errors.Wrap(err, "failed to create sqllite store")
	}
	result, err = engine.NewSQLiteDB(fmt.Sprintf("%s/pixel.sqlite", c.Store.Path))
	return result, errors.Wrap(err, "can't initialize data store")
}

func makeDirs(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil { // If path is already a directory, MkdirAll does nothing
			return errors.Wrapf(err, "can't make directory %s", dir)
		}
	}
	return nil
}
