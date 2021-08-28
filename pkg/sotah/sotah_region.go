package sotah

import (
	"encoding/json"
	"errors"
	"fmt"

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

func (rl RegionList) Names() []blizzardv2.RegionName {
	out := make([]blizzardv2.RegionName, len(rl))
	for i, region := range rl {
		out[i] = region.Name
	}

	return out
}

func (rl RegionList) FilterOut(names []blizzardv2.RegionName) RegionList {
	namesMap := map[blizzardv2.RegionName]struct{}{}
	for _, name := range names {
		namesMap[name] = struct{}{}
	}

	out := RegionList{}
	for _, region := range rl {
		if _, ok := namesMap[region.Name]; ok {
			continue
		}

		out = append(out, region)
	}

	return out
}

func NewRegion(data []byte) (Region, error) {
	out := Region{}
	if err := json.Unmarshal(data, &out); err != nil {
		return Region{}, err
	}

	return out, nil
}

type Region struct {
	Name     blizzardv2.RegionName `json:"name"`
	Hostname string                `json:"hostname"`
	Primary  bool                  `json:"primary"`
}

func (region Region) EncodeForStorage() ([]byte, error) {
	return json.Marshal(region)
}

func (region Region) String() string {
	return fmt.Sprintf(
		"name: %s, hostname: %s, primary: %t",
		region.Name,
		region.Hostname,
		region.Primary,
	)
}
