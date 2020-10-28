package areamaps

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func NewAreaMapsDatabase(dbDir string) (Database, error) {
	dbFilepath, err := databasePath(dbDir)
	if err != nil {
		return Database{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing area-maps database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{db}, nil
}

type Database struct {
	db *bolt.DB
}
