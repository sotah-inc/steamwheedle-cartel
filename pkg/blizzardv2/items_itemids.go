package blizzardv2

import "encoding/json"

func NewItemIds(data []byte) (ItemIds, error) {
	out := ItemIds{}
	if err := json.Unmarshal(data, &out); err != nil {
		return ItemIds{}, err
	}

	return out, nil
}

type ItemIds []ItemId

func (ids ItemIds) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(ids)
}

type ItemIdsMap map[ItemId]struct{}

func (idsMap ItemIdsMap) ToItemIds() ItemIds {
	out := make(ItemIds, len(idsMap))
	i := 0
	for id := range idsMap {
		out[i] = id

		i += 1
	}

	return out
}

func (ids ItemIds) ToUniqueMap() ItemIdsMap {
	out := ItemIdsMap{}
	for _, id := range ids {
		out[id] = struct{}{}
	}

	return out
}

func (ids ItemIds) Merge(providedIds ItemIds) ItemIds {
	results := ids.ToUniqueMap()
	for _, id := range providedIds {
		results[id] = struct{}{}
	}

	return results.ToItemIds()
}
