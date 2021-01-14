package stats

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func newTupleDatabase(dirPath string, tuple blizzardv2.RegionConnectedRealmTuple) (TupleDatabase, error) {
	dbFilepath := tupleDatabasePath(dirPath, tuple)
	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return TupleDatabase{}, err
	}

	return TupleDatabase{Database{db}, tuple}, nil
}

type TupleDatabase struct {
	Database
	tuple blizzardv2.RegionConnectedRealmTuple
}
