package sotah

import (
	"encoding/json"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/state"
)

const gcsAreaMapUrlFormat = "https://storage.googleapis.com/sotah-areamaps/%s/%d.jpg"

type AreaMapId int

type AreaMapMap map[AreaMapId]AreaMap

func (aMapMap AreaMapMap) SetUrls(version gameversions.GameVersion) AreaMapMap {
	out := AreaMapMap{}
	for id, aMap := range aMapMap {
		aMap.Url = aMap.GenerateUrl(version)

		out[id] = aMap
	}

	return out
}

func (aMapMap AreaMapMap) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(aMapMap)
}

func NewAreaMap(body []byte) (AreaMap, error) {
	areaMap := &AreaMap{}
	if err := json.Unmarshal(body, areaMap); err != nil {
		return AreaMap{}, err
	}

	return *areaMap, nil
}

type AreaMap struct {
	Id             AreaMapId   `json:"id"`
	State          state.State `json:"state"`
	Name           string      `json:"name"`
	NormalizedName string      `json:"normalized_name"`
	Url            string      `json:"-"`
}

func (aMap AreaMap) EncodeForStorage() ([]byte, error) {
	return json.Marshal(aMap)
}

func (aMap AreaMap) GenerateUrl(version gameversions.GameVersion) string {
	return fmt.Sprintf(gcsAreaMapUrlFormat, version, aMap.Id)
}

type AreaMapIdNameMap map[AreaMapId]string
