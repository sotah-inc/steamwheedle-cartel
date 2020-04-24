package database

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
		shards := func() LiveAuctionsDatabaseShards {
			out, ok := ladBases[tuple.RegionName]
			if !ok {
				return LiveAuctionsDatabaseShards{}
			}

			return out
		}()

		var err error
		shards[tuple.ConnectedRealmId], err = newLiveAuctionsDatabase(dirPath, tuple)
		if err != nil {
			return LiveAuctionsDatabases{}, err
		}

		ladBases[tuple.RegionName] = shards
	}

	return ladBases, nil
}

type LiveAuctionsDatabases map[blizzardv2.RegionName]LiveAuctionsDatabaseShards

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

type LiveAuctionsDatabaseShards map[blizzardv2.ConnectedRealmId]LiveAuctionsDatabase
