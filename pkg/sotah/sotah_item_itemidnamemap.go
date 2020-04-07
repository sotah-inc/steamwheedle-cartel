package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemIdNameMap(data []byte) (ItemIdNameMap, error) {
	gzipDecoded, err := util.GzipDecode(data)
	if err != nil {
		return ItemIdNameMap{}, err
	}

	var out ItemIdNameMap
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return ItemIdNameMap{}, err
	}

	return out, nil
}

type ItemIdNameMap map[blizzardv2.ItemId]locale.Mapping

func (idNameMap ItemIdNameMap) EncodeForDelivery() ([]byte, error) {
	jsonEncodedData, err := json.Marshal(idNameMap)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncodedData)
}

func (idNameMap ItemIdNameMap) ItemIds() []blizzardv2.ItemId {
	out := make([]blizzardv2.ItemId, len(idNameMap))
	i := 0
	for id := range idNameMap {
		out[i] = id

		i++
	}

	return out
}
