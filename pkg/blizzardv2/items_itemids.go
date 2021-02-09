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

func (ids ItemIds) Sub(providedIds ItemIds) ItemIds {
	providedIdsMap := providedIds.ToUniqueMap()
	out := ItemIds{}
	for _, id := range ids {
		if _, ok := providedIdsMap[id]; ok {
			continue
		}

		out = append(out, id)
	}

	return out
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

func (idsMap ItemIdsMap) Exists(id ItemId) bool {
	_, ok := idsMap[id]

	return ok
}
