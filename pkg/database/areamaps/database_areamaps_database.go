package areamaps

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func NewAreaMapsDatabase(dbDir string) (AreaMapsDatabase, error) {
	dbFilepath, err := AreaMapsDatabasePath(dbDir)
	if err != nil {
		return AreaMapsDatabase{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing area-maps database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return AreaMapsDatabase{}, err
	}

	return AreaMapsDatabase{db}, nil
}

type AreaMapsDatabase struct {
	db *bolt.DB
}
