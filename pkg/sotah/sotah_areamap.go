package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
)

type AreaMapId int

type AreaMapMap map[AreaMapId]AreaMap

func NewAreaMap(body []byte) (AreaMap, error) {
	areaMap := &AreaMap{}
	if err := json.Unmarshal(body, areaMap); err != nil {
		return AreaMap{}, err
	}

	return *areaMap, nil
}

type AreaMap struct {
	Id             AreaMapId
	State          state.State
	Name           string
	NormalizedName string
}

func (aMap AreaMap) EncodeForStorage() ([]byte, error) {
	return json.Marshal(aMap)
}
