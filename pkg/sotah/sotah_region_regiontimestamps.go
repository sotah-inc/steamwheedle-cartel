package sotah

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type ConnectedRealmTimestamps struct {
	Downloaded                 UnixTimestamp `json:"downloaded"`
	LiveAuctionsReceived       UnixTimestamp `json:"live_auctions_received"`
	PricelistHistoriesReceived UnixTimestamp `json:"pricelist_histories_received"`
}

func (timestamps ConnectedRealmTimestamps) IsZero() bool {
	return !timestamps.Downloaded.IsZero() &&
		!timestamps.LiveAuctionsReceived.IsZero() &&
		!timestamps.PricelistHistoriesReceived.IsZero()
}

func (timestamps ConnectedRealmTimestamps) Merge(in ConnectedRealmTimestamps) ConnectedRealmTimestamps {
	if !in.Downloaded.IsZero() {
		timestamps.Downloaded = in.Downloaded
	}

	if !in.LiveAuctionsReceived.IsZero() {
		timestamps.LiveAuctionsReceived = in.LiveAuctionsReceived
	}

	if !in.PricelistHistoriesReceived.IsZero() {
		timestamps.PricelistHistoriesReceived = in.PricelistHistoriesReceived
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

func (regionTimestamps RegionTimestamps) Exists(name blizzardv2.RegionName, id blizzardv2.ConnectedRealmId) bool {
	if _, ok := regionTimestamps[name]; !ok {
		return false
	}

	_, ok := regionTimestamps[name][id]

	return ok
}

func (regionTimestamps RegionTimestamps) resolve(
	name blizzardv2.RegionName,
	id blizzardv2.ConnectedRealmId,
) RegionTimestamps {
	if _, ok := regionTimestamps[name]; !ok {
		regionTimestamps[name] = map[blizzardv2.ConnectedRealmId]ConnectedRealmTimestamps{}
	}

	if _, ok := regionTimestamps[name][id]; !ok {
		regionTimestamps[name][id] = ConnectedRealmTimestamps{}
	}

	return regionTimestamps
}

func (regionTimestamps RegionTimestamps) SetDownloaded(
	name blizzardv2.RegionName,
	id blizzardv2.ConnectedRealmId,
	downloaded time.Time,
) RegionTimestamps {
	// resolving due to missing members
	out := regionTimestamps.resolve(name, id)

	// pushing the new time into the found member
	result := out[name][id]
	result.Downloaded = UnixTimestamp(downloaded.Unix())
	out[name][id] = result

	return out
}
