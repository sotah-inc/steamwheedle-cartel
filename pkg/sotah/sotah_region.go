package sotah

import (
	"errors"

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

type RegionComposite struct {
	Region
	ConnectedRealms []blizzardv2.ConnectedRealmResponse
}

func (region RegionComposite) ToTuples() []blizzardv2.RegionConnectedRealmTuple {
	out := make([]blizzardv2.RegionConnectedRealmTuple, len(region.ConnectedRealms))
	i := 0
	for _, response := range region.ConnectedRealms {
		out[i] = blizzardv2.RegionConnectedRealmTuple{
			RegionName:       region.Name,
			RegionHostname:   region.Hostname,
			ConnectedRealmId: response.Id,
		}
		i += 1
	}

	return out
}

type RegionComposites []RegionComposite

func (regions RegionComposites) TotalConnectedRealms() int {
	out := 0
	for _, region := range regions {
		out += len(region.ConnectedRealms)
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
