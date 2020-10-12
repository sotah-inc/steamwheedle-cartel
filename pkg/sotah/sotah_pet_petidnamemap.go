package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewPetIdNameMap(data []byte) (PetIdNameMap, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return PetIdNameMap{}, err
	}

	var out PetIdNameMap
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return PetIdNameMap{}, err
	}

	return out, nil
}

type PetIdNameMap map[blizzardv2.PetId]locale.Mapping

func (idNameMap PetIdNameMap) EncodeForDelivery() ([]byte, error) {
	jsonEncodedData, err := json.Marshal(idNameMap)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncodedData)
}

func (idNameMap PetIdNameMap) PetIds() []blizzardv2.PetId {
	out := make([]blizzardv2.PetId, len(idNameMap))
	i := 0
	for id := range idNameMap {
		out[i] = id

		i++
	}

	return out
}
