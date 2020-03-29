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
