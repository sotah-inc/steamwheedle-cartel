package sotah

import (
	"encoding/json"
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/util"
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

func (rl RegionList) GetRegion(name blizzard.RegionName) (Region, error) {
	for _, reg := range rl {
		if reg.Name == name {
			return reg, nil
		}
	}

	return Region{}, errors.New("failed to resolve region from name")
}

func (rl RegionList) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(rl)
	if err != nil {
		return []byte{}, err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncoded, nil
}

type Region struct {
	Name     blizzard.RegionName `json:"name"`
	Hostname string              `json:"hostname"`
	Primary  bool                `json:"primary"`
}
