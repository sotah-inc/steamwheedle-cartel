package liveauctions

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
)

func newLiveAuctionsDatabase(dirPath string, tuple blizzardv2.RegionConnectedRealmTuple) (LiveAuctionsDatabase, error) {
	dbFilepath := databasePath(dirPath, tuple)
	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return LiveAuctionsDatabase{}, err
	}

	return LiveAuctionsDatabase{db, tuple}, nil
}

type LiveAuctionsDatabase struct {
	db    *bolt.DB
	tuple blizzardv2.RegionConnectedRealmTuple
}
