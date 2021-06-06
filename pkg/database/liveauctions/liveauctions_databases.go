package liveauctions

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewDatabases(
	dirPath string,
	tuples []blizzardv2.VersionRegionConnectedRealmTuple,
) (Databases, error) {
	ladBases := Databases{}

	for _, tuple := range tuples {
		regionDatabases := func() RegionDatabases {
			out, ok := ladBases[tuple.Version]
			if !ok {
				return RegionDatabases{}
			}

			return out
		}()

		connectedRealmDatabases := func() ConnectedRealmDatabases {
			out, ok := regionDatabases[tuple.RegionName]
			if !ok {
				return ConnectedRealmDatabases{}
			}

			return out
		}()

		connectedRealmDatabase, err := newDatabase(dirPath, tuple)
		if err != nil {
			return Databases{}, err
		}

		connectedRealmDatabases[tuple.ConnectedRealmId] = connectedRealmDatabase
		regionDatabases[tuple.RegionName] = connectedRealmDatabases
		ladBases[tuple.Version] = regionDatabases
	}

	return ladBases, nil
}

type ConnectedRealmDatabases map[blizzardv2.ConnectedRealmId]Database

type RegionDatabases map[blizzardv2.RegionName]ConnectedRealmDatabases

type Databases map[gameversion.GameVersion]RegionDatabases

func (ladBases Databases) GetDatabase(
	tuple blizzardv2.VersionRegionConnectedRealmTuple,
) (Database, error) {
	regionDatabases, ok := ladBases[tuple.Version]
	if !ok {
		return Database{}, fmt.Errorf("shard not found for version %s", tuple.Version)
	}

	connectedRealmDatabases, ok := regionDatabases[tuple.RegionName]
	if !ok {
		return Database{}, fmt.Errorf("shard not found for region %s", tuple.RegionName)
	}

	connectedRealmDatabase, ok := connectedRealmDatabases[tuple.ConnectedRealmId]
	if !ok {
		return Database{}, fmt.Errorf("db not found for connected-realm %d", tuple.ConnectedRealmId)
	}

	return connectedRealmDatabase, nil
}
