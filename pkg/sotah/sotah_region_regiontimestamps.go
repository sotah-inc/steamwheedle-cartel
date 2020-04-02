package sotah

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type ConnectedRealmTimestamps struct {
	Downloaded time.Time
}

type RegionTimestamps map[blizzardv2.RegionName]map[blizzardv2.ConnectedRealmId]ConnectedRealmTimestamps

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
	result.Downloaded = downloaded
	out[name][id] = result

	return out
}
