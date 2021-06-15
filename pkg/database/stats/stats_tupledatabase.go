package stats

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func newTupleDatabase(
	dirPath string,
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (TupleDatabase, error) {
	db, err := newDatabase(tupleDatabasePath(dirPath, tuple))
	if err != nil {
		return TupleDatabase{}, err
	}

	return TupleDatabase{db, tuple}, nil
}

type TupleDatabase struct {
	Database
	tuple blizzardv2.RegionVersionConnectedRealmTuple
}
