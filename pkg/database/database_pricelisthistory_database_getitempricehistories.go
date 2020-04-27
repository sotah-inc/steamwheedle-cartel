package database

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetItemPriceHistoriesJob struct {
	Err          error
	Id           blizzardv2.ItemId
	PriceHistory sotah.PriceHistory
}

func (phdBase PricelistHistoryDatabase) GetItemPriceHistories(ids blizzardv2.ItemIds) chan GetItemPriceHistoriesJob {
	// drawing channels
	in := make(chan blizzardv2.ItemId)
	out := make(chan GetItemPriceHistoriesJob)

	// spinning up workers
	worker := func() {
		for id := range in {
			history, err := phdBase.getItemPriceHistory(id)
			if err != nil {
				out <- GetItemPriceHistoriesJob{
					Err:          err,
					Id:           id,
					PriceHistory: sotah.PriceHistory{},
				}

				continue
			}

			out <- GetItemPriceHistoriesJob{
				Err:          nil,
				Id:           id,
				PriceHistory: history,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// spinning it up
	go func() {
		for _, itemId := range ids {
			in <- itemId
		}

		close(in)
	}()

	return out

}

func (phdBase PricelistHistoryDatabase) getItemPriceHistory(id blizzardv2.ItemId) (sotah.PriceHistory, error) {
	out := sotah.PriceHistory{}

	err := phdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(pricelistHistoryBucketName(id))
		if bkt == nil {
			return nil
		}

		value := bkt.Get(pricelistHistoryKeyName())
		if value == nil {
			return nil
		}

		var err error
		out, err = sotah.NewPriceHistoryFromBytes(value)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.PriceHistory{}, err
	}

	return out, nil

}
