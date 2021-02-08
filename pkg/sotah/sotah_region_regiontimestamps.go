package sotah

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type RealmTimestamps map[blizzardv2.ConnectedRealmId]ConnectedRealmTimestamps

type RegionTimestamps map[blizzardv2.RegionName]RealmTimestamps

func (regionTimestamps RegionTimestamps) FindByRegionName(
	name blizzardv2.RegionName,
) (map[blizzardv2.ConnectedRealmId]ConnectedRealmTimestamps, error) {
	found, ok := regionTimestamps[name]
	if !ok {
		return nil, errors.New("failed to find region connected-realm timestamps")
	}

	return found, nil
}

func (regionTimestamps RegionTimestamps) IsZero() bool {
	for _, connectedRealmTimestamps := range regionTimestamps {
		for _, timestamps := range connectedRealmTimestamps {
			if !timestamps.IsZero() {
				return false
			}
		}
	}

	return true
}

func (regionTimestamps RegionTimestamps) Exists(tuple blizzardv2.RegionConnectedRealmTuple) bool {
	if _, ok := regionTimestamps[tuple.RegionName]; !ok {
		return false
	}

	_, ok := regionTimestamps[tuple.RegionName][tuple.ConnectedRealmId]

	return ok
}

func (regionTimestamps RegionTimestamps) resolve(
	tuple blizzardv2.RegionConnectedRealmTuple,
) RegionTimestamps {
	if _, ok := regionTimestamps[tuple.RegionName]; !ok {
		regionTimestamps[tuple.RegionName] = map[blizzardv2.ConnectedRealmId]ConnectedRealmTimestamps{}
	}

	if _, ok := regionTimestamps[tuple.RegionName][tuple.ConnectedRealmId]; !ok {
		regionTimestamps[tuple.RegionName][tuple.ConnectedRealmId] = ConnectedRealmTimestamps{}
	}

	return regionTimestamps
}

func (regionTimestamps RegionTimestamps) SetDownloaded(
	tuple blizzardv2.RegionConnectedRealmTuple,
	downloaded time.Time,
) RegionTimestamps {
	// resolving due to missing members
	out := regionTimestamps.resolve(tuple)

	// pushing the new time into the found member
	result := out[tuple.RegionName][tuple.ConnectedRealmId]
	result.Downloaded = UnixTimestamp(downloaded.Unix())
	out[tuple.RegionName][tuple.ConnectedRealmId] = result

	return out
}

func (regionTimestamps RegionTimestamps) SetLiveAuctionsReceived(
	tuple blizzardv2.RegionConnectedRealmTuple,
	liveAuctionsReceived time.Time,
) RegionTimestamps {
	// resolving due to missing members
	out := regionTimestamps.resolve(tuple)

	// pushing the new time into the found member
	result := out[tuple.RegionName][tuple.ConnectedRealmId]
	result.LiveAuctionsReceived = UnixTimestamp(liveAuctionsReceived.Unix())
	out[tuple.RegionName][tuple.ConnectedRealmId] = result

	return out
}

func (regionTimestamps RegionTimestamps) SetItemPricesReceived(
	tuple blizzardv2.RegionConnectedRealmTuple,
	itemPricesReceived time.Time,
) RegionTimestamps {
	// resolving due to missing members
	out := regionTimestamps.resolve(tuple)

	// pushing the new time into the found member
	result := out[tuple.RegionName][tuple.ConnectedRealmId]
	result.ItemPricesReceived = UnixTimestamp(itemPricesReceived.Unix())
	out[tuple.RegionName][tuple.ConnectedRealmId] = result

	return out
}

func (regionTimestamps RegionTimestamps) SetRecipePricesReceived(
	tuple blizzardv2.RegionConnectedRealmTuple,
	recipePricesReceived time.Time,
) RegionTimestamps {
	// resolving due to missing members
	out := regionTimestamps.resolve(tuple)

	// pushing the new time into the found member
	result := out[tuple.RegionName][tuple.ConnectedRealmId]
	result.RecipePricesReceived = UnixTimestamp(recipePricesReceived.Unix())
	out[tuple.RegionName][tuple.ConnectedRealmId] = result

	return out
}

func (regionTimestamps RegionTimestamps) SetStatsReceived(
	tuple blizzardv2.RegionConnectedRealmTuple,
	statsReceived time.Time,
) RegionTimestamps {
	// resolving due to missing members
	out := regionTimestamps.resolve(tuple)

	// pushing the new time into the found member
	result := out[tuple.RegionName][tuple.ConnectedRealmId]
	result.RecipePricesReceived = UnixTimestamp(statsReceived.Unix())
	out[tuple.RegionName][tuple.ConnectedRealmId] = result

	return out
}
