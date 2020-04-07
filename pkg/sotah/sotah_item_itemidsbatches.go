package sotah

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

func NewItemIdsBatches(ids []blizzardv2.ItemId, batchSize int) ItemIdBatches {
	batches := ItemIdBatches{}
	for i, id := range ids {
		key := (i - (i % batchSize)) / batchSize
		batch := func() []blizzardv2.ItemId {
			out, ok := batches[key]
			if !ok {
				return []blizzardv2.ItemId{}
			}

			return out
		}()
		batch = append(batch, id)

		batches[key] = batch
	}

	return batches
}

type ItemIdBatches map[int][]blizzardv2.ItemId
