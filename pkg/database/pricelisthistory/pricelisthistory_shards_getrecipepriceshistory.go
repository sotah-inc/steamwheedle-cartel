package pricelisthistory

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetRecipePricesHistoryJob struct {
	Err          error
	PriceHistory sotah.RecipePriceHistory
}

func (job GetRecipePricesHistoryJob) ToLogrusFields(id blizzardv2.RecipeId) logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"recipe": id,
	}
}

func (shards DatabaseShards) GetRecipePricesHistory(
	id blizzardv2.RecipeId,
	lowerBounds sotah.UnixTimestamp,
	upperBounds sotah.UnixTimestamp,
) (sotah.RecipePriceHistory, error) {
	// establish channels
	in := make(chan Database)
	out := make(chan GetRecipePricesHistoryJob)

	// spinning up workers for querying price-histories
	worker := func() {
		for phdBase := range in {
			recipePrices, err := phdBase.getRecipePrices(id)
			if err != nil {
				out <- GetRecipePricesHistoryJob{
					Err:          err,
					PriceHistory: nil,
				}

				continue
			}

			out <- GetRecipePricesHistoryJob{
				Err:          nil,
				PriceHistory: sotah.RecipePriceHistory{phdBase.targetTimestamp: recipePrices},
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	go func() {
		for _, phdBase := range shards {
			in <- phdBase
		}

		close(in)
	}()

	rpHistory := sotah.RecipePriceHistory{}
	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields(id)).Error("failed to fetch price history for recipe")

			return sotah.RecipePriceHistory{}, job.Err
		}

		rpHistory = rpHistory.Merge(job.PriceHistory.Between(lowerBounds, upperBounds))
	}

	return rpHistory, nil
}
