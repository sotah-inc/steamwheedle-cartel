package liveauctions

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewDatabases(
	dirPath string,
	tuples []blizzardv2.RegionConnectedRealmTuple,
) (Databases, error) {
	ladBases := Databases{}

	for _, tuple := range tuples {
		connectedRealmDatabases := func() map[blizzardv2.ConnectedRealmId]Database {
			out, ok := ladBases[tuple.RegionName]
			if !ok {
				return map[blizzardv2.ConnectedRealmId]Database{}
			}

			return out
		}()

		var err error
		connectedRealmDatabases[tuple.ConnectedRealmId], err = newDatabase(dirPath, tuple)
		if err != nil {
			return Databases{}, err
		}

		ladBases[tuple.RegionName] = connectedRealmDatabases
	}

	return ladBases, nil
}

type Databases map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]Database

func (ladBases Databases) GetDatabase(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (Database, error) {
	shards, ok := ladBases[tuple.RegionName]
	if !ok {
		return Database{}, fmt.Errorf("shard not found for region %s", tuple.RegionName)
	}

	db, ok := shards[tuple.ConnectedRealmId]
	if !ok {
		return Database{}, fmt.Errorf("db not found for connected-realm %d", tuple.ConnectedRealmId)
	}

	return db, nil
}
