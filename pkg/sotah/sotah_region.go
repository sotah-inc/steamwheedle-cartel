package sotah

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type RegionList []Region

func (rl RegionList) GetPrimaryRegion() (Region, error) {
	for _, reg := range rl {
		if reg.Primary {
			return reg, nil
		}
	}

	return Region{}, errors.New("could not find primary region")
}

func (rl RegionList) GetRegion(name blizzardv2.RegionName) (Region, error) {
	for _, reg := range rl {
		if reg.Name == name {
			return reg, nil
		}
	}

	return Region{}, errors.New("failed to resolve region from name")
}

type Region struct {
	Name     blizzardv2.RegionName `json:"name"`
	Hostname string                `json:"hostname"`
	Primary  bool                  `json:"primary"`
}

type RealmComposite struct {
	ConnectedRealmResponse blizzardv2.ConnectedRealmResponse
	ModificationDates      struct {
		Downloaded                 time.Time
		LiveAuctionsReceived       time.Time
		PricelistHistoriesReceived time.Time
	}
}

type RegionComposite struct {
	ConfigRegion             Region
	ConnectedRealmComposites []RealmComposite
}

func (region RegionComposite) ToTuples() []blizzardv2.RegionConnectedRealmTuple {
	out := make([]blizzardv2.RegionConnectedRealmTuple, len(region.ConnectedRealmComposites))
	i := 0
	for _, composite := range region.ConnectedRealmComposites {
		out[i] = blizzardv2.RegionConnectedRealmTuple{
			RegionName:       region.ConfigRegion.Name,
			RegionHostname:   region.ConfigRegion.Hostname,
			ConnectedRealmId: composite.ConnectedRealmResponse.Id,
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

func (regions RegionComposites) ToTuples() []blizzardv2.RegionConnectedRealmTuple {
	out := make([]blizzardv2.RegionConnectedRealmTuple, regions.TotalConnectedRealms())
	i := 0
	for _, region := range regions {
		for _, tuple := range region.ToTuples() {
			out[i] = tuple
			i += 1
		}
	}

	return out
}
