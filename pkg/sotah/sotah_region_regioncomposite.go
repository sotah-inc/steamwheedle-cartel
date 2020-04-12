package sotah

import (
	"encoding/json"
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type RealmComposite struct {
	ConnectedRealmResponse blizzardv2.ConnectedRealmResponse `json:"connected_realm"`
	ModificationDates      ConnectedRealmTimestamps          `json:"modification_dates"`
}

type RegionComposite struct {
	ConfigRegion             Region           `json:"config_region"`
	ConnectedRealmComposites []RealmComposite `json:"connected_realms"`
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
			LastModified:   composite.ModificationDates.Downloaded.Time,
		}
		i += 1
	}

	return out
}

func (region RegionComposite) EncodeForDelivery() ([]byte, error) {
	jsonEncoded, err := json.Marshal(region)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

type RegionComposites []RegionComposite

func (regions RegionComposites) FindByRegionName(
	name blizzardv2.RegionName,
) (RegionComposite, error) {
	for _, region := range regions {
		if region.ConfigRegion.Name == name {
			return region, nil
		}
	}

	return RegionComposite{}, errors.New("failed to resolve connected-realms")
}

func (regions RegionComposites) RegionRealmExists(
	name blizzardv2.RegionName,
	slug blizzardv2.RealmSlug,
) bool {
	for _, region := range regions {
		if region.ConfigRegion.Name != name {
			continue
		}

		for _, connectedRealm := range region.ConnectedRealmComposites {
			for _, realm := range connectedRealm.ConnectedRealmResponse.Realms {
				if realm.Slug == slug {
					return true
				}
			}
		}
	}

	return false
}

func (regions RegionComposites) FindConnectedRealmTimestamps(
	name blizzardv2.RegionName,
	id blizzardv2.ConnectedRealmId,
) (ConnectedRealmTimestamps, error) {
	for _, region := range regions {
		if region.ConfigRegion.Name != name {
			continue
		}

		for _, connectedRealm := range region.ConnectedRealmComposites {
			if connectedRealm.ConnectedRealmResponse.Id == id {
				return connectedRealm.ModificationDates, nil
			}
		}
	}

	return ConnectedRealmTimestamps{}, errors.New("failed to resolve realm-timestamps")
}

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

func (regions RegionComposites) ToList() RegionList {
	out := make(RegionList, len(regions))
	for i, region := range regions {
		out[i] = region.ConfigRegion
	}

	return out
}
