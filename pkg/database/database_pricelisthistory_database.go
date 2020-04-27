package database

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func newPricelistHistoryDatabase(
	dbFilepath string,
	targetTimestamp sotah.UnixTimestamp,
) (PricelistHistoryDatabase, error) {
	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return PricelistHistoryDatabase{}, err
	}

	return PricelistHistoryDatabase{
		db:              db,
		targetTimestamp: targetTimestamp,
	}, nil
}

type PricelistHistoryDatabase struct {
	db              *bolt.DB
	targetTimestamp sotah.UnixTimestamp
}
