package liveauctions

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
)

func newDatabase(dirPath string, tuple blizzardv2.RegionConnectedRealmTuple) (Database, error) {
	dbFilepath := databasePath(dirPath, tuple)
	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{db, tuple}, nil
}

type Database struct {
	db    *bolt.DB
	tuple blizzardv2.RegionConnectedRealmTuple
}
