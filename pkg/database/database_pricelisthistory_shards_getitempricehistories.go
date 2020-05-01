package database

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetItemPriceHistoriesJob struct {
	Err          error
	Id           blizzardv2.ItemId
	PriceHistory sotah.PriceHistory
}

func (job GetItemPriceHistoriesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"item":  job.Id,
	}
}

func (shards PricelistHistoryDatabaseShards) GetItemPriceHistories(
	ids blizzardv2.ItemIds,
	lowerBounds sotah.UnixTimestamp,
	upperBounds sotah.UnixTimestamp,
) (sotah.ItemPriceHistories, error) {
	// establish channels
	in := make(chan blizzardv2.ItemId)
	out := make(chan GetItemPriceHistoriesJob)

	// spinning up workers for querying price-histories
	worker := func() {
		for id := range in {
			receivedHistory, err := shards.GetPriceHistory(id, lowerBounds, upperBounds)
			if err != nil {
				out <- GetItemPriceHistoriesJob{
					Err:          err,
					Id:           id,
					PriceHistory: nil,
				}

				continue
			}

			out <- GetItemPriceHistoriesJob{
				Err:          nil,
				Id:           id,
				PriceHistory: receivedHistory,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	itemPriceHistories := sotah.ItemPriceHistories{}
	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to fetch price history for item")

			return sotah.ItemPriceHistories{}, job.Err
		}

		priceHistory := func() sotah.PriceHistory {
			found, ok := itemPriceHistories[job.Id]
			if !ok {
				return sotah.PriceHistory{}
			}

			return found
		}()

		itemPriceHistories[job.Id] = priceHistory.Merge(job.PriceHistory)
	}

	return itemPriceHistories, nil
}
