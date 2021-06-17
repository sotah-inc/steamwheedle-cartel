package pricelisthistory

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type getItemsMarketPricesJob struct {
	err         error
	id          blizzardv2.ItemId
	marketPrice float64
}

func (job getItemsMarketPricesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.err.Error(),
		"item":  job.id,
	}
}

func (phdBases *Databases) GetItemsMarketPrice(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	ids blizzardv2.ItemIds,
) (map[blizzardv2.ItemId]float64, error) {
	// resolving shards for this tuple
	shards, err := phdBases.GetShards(tuple)
	if err != nil {
		return map[blizzardv2.ItemId]float64{}, err
	}

	// resolving latest pricelist-history db
	phdBase, err := shards.Latest()
	if err != nil {
		return map[blizzardv2.ItemId]float64{}, err
	}

	// spinning up workers for receiving item market-prices
	in := make(chan blizzardv2.ItemId)
	out := make(chan getItemsMarketPricesJob)
	worker := func() {
		for id := range in {
			marketPrice, err := phdBase.getItemMarketPrice(id)
			if err != nil {
				out <- getItemsMarketPricesJob{
					err:         err,
					id:          id,
					marketPrice: 0,
				}

				continue
			}

			out <- getItemsMarketPricesJob{
				err:         nil,
				id:          id,
				marketPrice: marketPrice,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	// filling in default values
	results := map[blizzardv2.ItemId]float64{}
	for _, id := range ids {
		results[id] = 0
	}

	// waiting for it to drain out
	for job := range out {
		if job.err != nil {
			return map[blizzardv2.ItemId]float64{}, job.err
		}

		results[job.id] = job.marketPrice
	}

	return results, nil
}
