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

type TupleDatabaseShards map[blizzardv2.ConnectedRealmId]TupleDatabase

type TupleDatabases map[blizzardv2.RegionName]TupleDatabaseShards

func (tBases TupleDatabases) GetTupleDatabase(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (TupleDatabase, error) {
	shard, err := tBases.GetRegionShard(tuple.RegionName)
	if err != nil {
		return TupleDatabase{}, err
	}

	db, ok := shard[tuple.ConnectedRealmId]
	if !ok {
		return TupleDatabase{}, fmt.Errorf("db not found for connected-realm %d", tuple.ConnectedRealmId)
	}

	return db, nil
}

func (tBases TupleDatabases) GetRegionShard(
	regionName blizzardv2.RegionName,
) (TupleDatabaseShards, error) {
	shards, ok := tBases[regionName]
	if !ok {
		return TupleDatabaseShards{}, fmt.Errorf("shard not found for region %s", regionName)
	}

	return shards, nil
}
