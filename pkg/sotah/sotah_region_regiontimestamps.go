package sotah

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type ConnectedRealmTimestamps struct {
	Downloaded               UnixTimestamp `json:"downloaded"`
	LiveAuctionsReceived     UnixTimestamp `json:"live_auctions_received"`
	PricelistHistoryReceived UnixTimestamp `json:"pricelist_history_received"`
}

func (timestamps ConnectedRealmTimestamps) IsZero() bool {
	return !timestamps.Downloaded.IsZero() &&
		!timestamps.LiveAuctionsReceived.IsZero() &&
		!timestamps.PricelistHistoryReceived.IsZero()
}

func (timestamps ConnectedRealmTimestamps) Merge(in ConnectedRealmTimestamps) ConnectedRealmTimestamps {
	if !in.Downloaded.IsZero() {
		timestamps.Downloaded = in.Downloaded
	}

	if !in.LiveAuctionsReceived.IsZero() {
		timestamps.LiveAuctionsReceived = in.LiveAuctionsReceived
	}

	if !in.PricelistHistoryReceived.IsZero() {
		timestamps.PricelistHistoryReceived = in.PricelistHistoryReceived
	}

	return timestamps
}

type RegionTimestamps map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]ConnectedRealmTimestamps

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

func (regionTimestamps RegionTimestamps) SetPricelistHistoryReceived(
	tuple blizzardv2.RegionConnectedRealmTuple,
	liveAuctionsReceived time.Time,
) RegionTimestamps {
	// resolving due to missing members
	out := regionTimestamps.resolve(tuple)

	// pushing the new time into the found member
	result := out[tuple.RegionName][tuple.ConnectedRealmId]
	result.PricelistHistoryReceived = UnixTimestamp(liveAuctionsReceived.Unix())
	out[tuple.RegionName][tuple.ConnectedRealmId] = result

	return out
}
