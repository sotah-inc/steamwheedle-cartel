package stats

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func NewTupleDatabases(
	dirPath string,
	tuples blizzardv2.RegionVersionConnectedRealmTuples,
) (TupleDatabases, error) {
	tBases := make(TupleDatabases, len(tuples))

	for i, tuple := range tuples {
		logging.WithFields(logrus.Fields{
			"version": tuple.Version,
			"region":  tuple.RegionName,
			"realm":   tuple.ConnectedRealmId,
		}).Info("acquiring new database")

		var err error
		tBases[i], err = newTupleDatabase(dirPath, tuple)
		if err != nil {
			return TupleDatabases{}, err
		}
	}

	return tBases, nil
}

type TupleDatabases []TupleDatabase

func (tBases TupleDatabases) GetTupleDatabase(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (TupleDatabase, error) {
	for _, db := range tBases {
		if db.tuple.RegionName != tuple.RegionName {
			continue
		}

		if db.tuple.Version != tuple.Version {
			continue
		}

		if db.tuple.ConnectedRealmId != tuple.ConnectedRealmId {
			continue
		}

		return db, nil
	}

	return TupleDatabase{}, fmt.Errorf(
		"failed to resolve tuple database with tuple: %s",
		tuple.String(),
	)
}

func (tBases TupleDatabases) GetTupleDatabasesByRegionName(
	name blizzardv2.RegionName,
) TupleDatabases {
	out := TupleDatabases{}

	for _, db := range tBases {
		if db.tuple.RegionName != name {
			continue
		}

		out = append(out, db)
	}

	return out
}
