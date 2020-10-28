package liveauctions

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewLiveAuctionsDatabases(
	dirPath string,
	tuples []blizzardv2.RegionConnectedRealmTuple,
) (LiveAuctionsDatabases, error) {
	ladBases := LiveAuctionsDatabases{}

	for _, tuple := range tuples {
		connectedRealmDatabases := func() map[blizzardv2.ConnectedRealmId]LiveAuctionsDatabase {
			out, ok := ladBases[tuple.RegionName]
			if !ok {
				return map[blizzardv2.ConnectedRealmId]LiveAuctionsDatabase{}
			}

			return out
		}()

		var err error
		connectedRealmDatabases[tuple.ConnectedRealmId], err = newLiveAuctionsDatabase(dirPath, tuple)
		if err != nil {
			return LiveAuctionsDatabases{}, err
		}

		ladBases[tuple.RegionName] = connectedRealmDatabases
	}

	return ladBases, nil
}

type LiveAuctionsDatabases map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]LiveAuctionsDatabase

func (ladBases LiveAuctionsDatabases) GetDatabase(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (LiveAuctionsDatabase, error) {
	shards, ok := ladBases[tuple.RegionName]
	if !ok {
		return LiveAuctionsDatabase{}, fmt.Errorf("shard not found for region %s", tuple.RegionName)
	}

	db, ok := shards[tuple.ConnectedRealmId]
	if !ok {
		return LiveAuctionsDatabase{}, fmt.Errorf("db not found for connected-realm %d", tuple.ConnectedRealmId)
	}

	return db, nil
}
