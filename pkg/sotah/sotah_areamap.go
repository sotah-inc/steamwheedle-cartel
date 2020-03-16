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

func (aMap AreaMap) GenerateUrl(version gameversions.GameVersion) string {
	return fmt.Sprintf(gcsAreaMapUrlFormat, version, aMap.Id)
}

type AreaMapIdNameMap map[AreaMapId]string
