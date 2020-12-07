package pricelisthistory

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetRecipePriceHistoriesJob struct {
	Err          error
	Id           blizzardv2.RecipeId
	PriceHistory sotah.RecipePriceHistory
}

func (job GetRecipePriceHistoriesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"recipe": job.Id,
	}
}

func (shards DatabaseShards) GetRecipePriceHistories(
	ids blizzardv2.RecipeIds,
	lowerBounds sotah.UnixTimestamp,
	upperBounds sotah.UnixTimestamp,
) (sotah.RecipePriceHistories, error) {
	// establish channels
	in := make(chan blizzardv2.RecipeId)
	out := make(chan GetRecipePriceHistoriesJob)

	// spinning up workers for querying price-histories
	worker := func() {
		for id := range in {
			receivedHistory, err := shards.GetRecipePricesHistory(id, lowerBounds, upperBounds)
			if err != nil {
				out <- GetRecipePriceHistoriesJob{
					Err:          err,
					Id:           id,
					PriceHistory: nil,
				}

				continue
			}

			out <- GetRecipePriceHistoriesJob{
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

	recipePriceHistories := sotah.RecipePriceHistories{}
	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to fetch price history for recipe")

			return sotah.RecipePriceHistories{}, job.Err
		}

		priceHistory := func() sotah.RecipePriceHistory {
			found, ok := recipePriceHistories[job.Id]
			if !ok {
				return sotah.RecipePriceHistory{}
			}

			return found
		}()

		recipePriceHistories[job.Id] = priceHistory.Merge(job.PriceHistory)
	}

	return recipePriceHistories, nil
}
