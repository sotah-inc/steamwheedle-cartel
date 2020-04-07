package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type IconIdsMap map[IconName]blizzardv2.ItemIds

func (iMap IconIdsMap) Append(name IconName, id blizzardv2.ItemId) IconIdsMap {
	if _, ok := iMap[name]; !ok {
		iMap[name] = blizzardv2.ItemIds{}
	}

	iMap[name] = append(iMap[name], id)

	return iMap
}

func NewIconItemsTuplesBatches(iconIdsMap IconIdsMap, batchSize int) IconItemsTuplesBatches {
	batches := IconItemsTuplesBatches{}
	i := 0
	for iconName, itemIds := range iconIdsMap {
		key := (i - (i % batchSize)) / batchSize
		batch := func() IconItemsTuples {
			out, ok := batches[key]
			if !ok {
				return IconItemsTuples{}
			}

			return out
		}()
		batch = append(batch, IconItemsTuple{Name: iconName, Ids: itemIds})

		batches[key] = batch

		i += 1
	}

	return batches
}

type IconItemsTuplesBatches map[int]IconItemsTuples

type IconItemsTuples []IconItemsTuple

func (d IconItemsTuples) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(d)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

type IconItemsTuple struct {
	Name IconName
	Ids  []blizzardv2.ItemId
}
