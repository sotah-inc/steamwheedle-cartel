package stats

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewTupleDatabases(
	dirPath string,
	tuples []blizzardv2.RegionConnectedRealmTuple,
) (TupleDatabases, error) {
	tBases := TupleDatabases{}

	for _, tuple := range tuples {
		connectedRealmDatabases := func() map[blizzardv2.ConnectedRealmId]TupleDatabase {
			out, ok := tBases[tuple.RegionName]
			if !ok {
				return map[blizzardv2.ConnectedRealmId]TupleDatabase{}
			}

			return out
		}()

		var err error
		connectedRealmDatabases[tuple.ConnectedRealmId], err = newTupleDatabase(dirPath, tuple)
		if err != nil {
			return TupleDatabases{}, err
		}

		tBases[tuple.RegionName] = connectedRealmDatabases
	}

	return tBases, nil
}

type TupleDatabases map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]TupleDatabase

func (tBases TupleDatabases) GetDatabase(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (TupleDatabase, error) {
	shards, ok := tBases[tuple.RegionName]
	if !ok {
		return TupleDatabase{}, fmt.Errorf("shard not found for region %s", tuple.RegionName)
	}

	db, ok := shards[tuple.ConnectedRealmId]
	if !ok {
		return TupleDatabase{}, fmt.Errorf("db not found for connected-realm %d", tuple.ConnectedRealmId)
	}

	return db, nil
}
