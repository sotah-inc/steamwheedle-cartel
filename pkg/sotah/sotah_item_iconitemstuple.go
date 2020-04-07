package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewIconItemsTuplesBatches(iconIdsMap map[string][]blizzardv2.ItemId, batchSize int) IconItemsTuplesBatches {
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
	Name string
	Ids  []blizzardv2.ItemId
}
