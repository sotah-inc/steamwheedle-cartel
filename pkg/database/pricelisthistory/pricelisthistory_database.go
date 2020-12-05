package pricelisthistory

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func newDatabase(
	dbFilepath string,
	targetTimestamp sotah.UnixTimestamp,
) (Database, error) {
	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{
		db:              db,
		targetTimestamp: targetTimestamp,
	}, nil
}

type Database struct {
	db              *bolt.DB
	targetTimestamp sotah.UnixTimestamp
}
