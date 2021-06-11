package liveauctions

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewDatabases(
	dirPath string,
	tuples []blizzardv2.RegionVersionConnectedRealmTuple,
) (Databases, error) {
	ladBases := Databases{}

	for _, tuple := range tuples {
		versionDatabases := func() VersionDatabases {
			out, ok := ladBases[tuple.RegionName]
			if !ok {
				return VersionDatabases{}
			}

			return out
		}()

		connectedRealmDatabases := func() ConnectedRealmDatabases {
			out, ok := versionDatabases[tuple.Version]
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
		versionDatabases[tuple.Version] = connectedRealmDatabases
		ladBases[tuple.RegionName] = versionDatabases
	}

	return ladBases, nil
}

type ConnectedRealmDatabases map[blizzardv2.ConnectedRealmId]Database

type VersionDatabases map[gameversion.GameVersion]ConnectedRealmDatabases

type Databases map[blizzardv2.RegionName]VersionDatabases

func (ladBases Databases) GetDatabase(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (Database, error) {
	regionDatabases, ok := ladBases[tuple.RegionName]
	if !ok {
		return Database{}, fmt.Errorf("shard not found for region %s", tuple.Version)
	}

	connectedRealmDatabases, ok := regionDatabases[tuple.Version]
	if !ok {
		return Database{}, fmt.Errorf("shard not found for version %s", tuple.RegionName)
	}

	connectedRealmDatabase, ok := connectedRealmDatabases[tuple.ConnectedRealmId]
	if !ok {
		return Database{}, fmt.Errorf("db not found for connected-realm %d", tuple.ConnectedRealmId)
	}

	return connectedRealmDatabase, nil
}
