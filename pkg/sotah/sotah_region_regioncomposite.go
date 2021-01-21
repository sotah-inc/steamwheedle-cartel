package sotah

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type RealmComposites []RealmComposite

func (comps RealmComposites) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(comps)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func NewRealmComposite(
	realmWhitelist blizzardv2.RealmSlugs,
	res blizzardv2.ConnectedRealmResponse,
) RealmComposite {
	if len(realmWhitelist) > 0 {
		res.Realms = res.Realms.FilterIn(realmWhitelist)
	}

	return RealmComposite{
		ConnectedRealmResponse: res,
		ModificationDates:      ConnectedRealmTimestamps{},
	}
}

type RealmComposite struct {
	ConnectedRealmResponse blizzardv2.ConnectedRealmResponse `json:"connected_realm"`
	ModificationDates      ConnectedRealmTimestamps          `json:"modification_dates"`
}

func (composite RealmComposite) IsZero() bool {
	return len(composite.ConnectedRealmResponse.Realms) == 0
}

type RegionComposite struct {
	ConfigRegion             Region          `json:"config_region"`
	ConnectedRealmComposites RealmComposites `json:"connected_realms"`
}

func (region RegionComposite) ToDownloadTuples() []blizzardv2.DownloadConnectedRealmTuple {
	out := make([]blizzardv2.DownloadConnectedRealmTuple, len(region.ConnectedRealmComposites))
	i := 0
	for _, composite := range region.ConnectedRealmComposites {
		out[i] = blizzardv2.DownloadConnectedRealmTuple{
			LoadConnectedRealmTuple: blizzardv2.LoadConnectedRealmTuple{
				RegionConnectedRealmTuple: blizzardv2.RegionConnectedRealmTuple{
					RegionName:       region.ConfigRegion.Name,
					ConnectedRealmId: composite.ConnectedRealmResponse.Id,
				},
				LastModified: composite.ModificationDates.Downloaded.Time(),
			},
			RegionHostname: region.ConfigRegion.Hostname,
		}
		i += 1
	}

	return out
}

func (region RegionComposite) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(region)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

type RegionComposites []RegionComposite

func (regions RegionComposites) EncodeForDelivery() ([]byte, error) {
	jsonEncoded, err := json.Marshal(regions)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

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

func (regions RegionComposites) FindConnectedRealm(
	tuple blizzardv2.RegionRealmTuple,
) (RealmComposite, error) {
	for _, region := range regions {
		if region.ConfigRegion.Name != tuple.RegionName {
			continue
		}

		for _, connectedRealm := range region.ConnectedRealmComposites {
			for _, realm := range connectedRealm.ConnectedRealmResponse.Realms {
				if realm.Slug != tuple.RealmSlug {
					continue
				}

				return connectedRealm, nil
			}
		}
	}

	return RealmComposite{}, errors.New("failed to find connected-realm")
}

func (regions RegionComposites) RegionConnectedRealmExists(
	name blizzardv2.RegionName,
	id blizzardv2.ConnectedRealmId,
) bool {
	for _, region := range regions {
		if region.ConfigRegion.Name != name {
			continue
		}

		for _, connectedRealm := range region.ConnectedRealmComposites {
			if connectedRealm.ConnectedRealmResponse.Id == id {
				return true
			}
		}
	}

	return false
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

func (regions RegionComposites) ToTuples() blizzardv2.RegionConnectedRealmTuples {
	out := make(blizzardv2.RegionConnectedRealmTuples, regions.TotalConnectedRealms())
	i := 0
	for _, region := range regions {
		for _, tuple := range region.ToDownloadTuples() {
			out[i] = tuple.RegionConnectedRealmTuple
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

			tuple := blizzardv2.RegionConnectedRealmTuple{
				RegionName:       regionName,
				ConnectedRealmId: connectedRealmId,
			}
			if !timestamps.Exists(tuple) {
				continue
			}

			regions[i].ConnectedRealmComposites[j].ModificationDates = connectedRealm.ModificationDates.Merge( // nolint:lll
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
