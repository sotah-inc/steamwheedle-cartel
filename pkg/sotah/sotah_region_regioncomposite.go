package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type RealmComposite struct {
	ConnectedRealmResponse blizzardv2.ConnectedRealmResponse
	ModificationDates      ConnectedRealmTimestamps
}

type RegionComposite struct {
	ConfigRegion             Region
	ConnectedRealmComposites []RealmComposite
}

func (region RegionComposite) ToDownloadTuples() []blizzardv2.DownloadConnectedRealmTuple {
	out := make([]blizzardv2.DownloadConnectedRealmTuple, len(region.ConnectedRealmComposites))
	i := 0
	for _, composite := range region.ConnectedRealmComposites {
		out[i] = blizzardv2.DownloadConnectedRealmTuple{
			RegionConnectedRealmTuple: blizzardv2.RegionConnectedRealmTuple{
				RegionName:       region.ConfigRegion.Name,
				ConnectedRealmId: composite.ConnectedRealmResponse.Id,
			},
			RegionHostname: region.ConfigRegion.Hostname,
			LastModified:   composite.ModificationDates.Downloaded,
		}
		i += 1
	}

	return out
}

type RegionComposites []RegionComposite

func (regions RegionComposites) TotalConnectedRealms() int {
	out := 0
	for _, region := range regions {
		out += len(region.ConnectedRealmComposites)
	}

	return out
}

func (regions RegionComposites) ToDownloadTuples() []blizzardv2.DownloadConnectedRealmTuple {
	out := make([]blizzardv2.DownloadConnectedRealmTuple, regions.TotalConnectedRealms())
	i := 0
	for _, region := range regions {
		for _, tuple := range region.ToDownloadTuples() {
			out[i] = tuple
			i += 1
		}
	}

	return out
}

func (regions RegionComposites) Receive(timestamps RegionTimestamps) RegionComposites {
	for i, region := range regions {
		regionName := region.ConfigRegion.Name

		for j, connectedRealm := range region.ConnectedRealmComposites {
			connectedRealmId := connectedRealm.ConnectedRealmResponse.Id

			if !timestamps.Exists(regionName, connectedRealmId) {
				continue
			}

			regions[i].ConnectedRealmComposites[j].ModificationDates = connectedRealm.ModificationDates.Merge(
				timestamps[regionName][connectedRealmId],
			)
		}
	}

	return regions
}
