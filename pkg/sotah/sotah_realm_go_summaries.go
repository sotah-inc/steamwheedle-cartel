package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type RegionRealmSummaryTuples []RegionRealmSummaryTuple

func (tuples RegionRealmSummaryTuples) ItemIds() []blizzardv2.ItemId {
	itemIdsMap := ItemIdsMap{}
	for _, tuple := range tuples {
		for _, id := range tuple.ItemIds {
			itemIdsMap[blizzardv2.ItemId(id)] = struct{}{}
		}
	}

	out := make([]blizzardv2.ItemId, len(itemIdsMap))
	for id := range itemIdsMap {
		out = append(out, id)
	}

	return out
}

func (tuples RegionRealmSummaryTuples) RegionRealmTuples() RegionRealmTuples {
	out := make(RegionRealmTuples, len(tuples))
	for i, tuple := range tuples {
		out[i] = tuple.RegionRealmTuple
	}

	return out
}

func NewRegionRealmSummaryTuple(data string) (RegionRealmSummaryTuple, error) {
	var out RegionRealmSummaryTuple
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return RegionRealmSummaryTuple{}, err
	}

	return out, nil
}

type RegionRealmSummaryTuple struct {
	RegionRealmTimestampTuple
	ItemIds []int `json:"item_ids"`
}

func (tuple RegionRealmSummaryTuple) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(tuple)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}
