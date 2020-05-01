package database

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetPriceHistoryJob struct {
	Err          error
	PriceHistory sotah.PriceHistory
}

func (job GetPriceHistoryJob) ToLogrusFields(id blizzardv2.ItemId) logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"item":  id,
	}
}

func (shards PricelistHistoryDatabaseShards) GetPriceHistory(
	id blizzardv2.ItemId,
	lowerBounds sotah.UnixTimestamp,
	upperBounds sotah.UnixTimestamp,
) (sotah.PriceHistory, error) {
	// establish channels
	in := make(chan PricelistHistoryDatabase)
	out := make(chan GetPriceHistoryJob)

	// spinning up workers for querying price-histories
	worker := func() {
		for phdBase := range in {
			receivedHistory, err := phdBase.getItemPriceHistory(id)
			if err != nil {
				out <- GetPriceHistoryJob{
					Err:          err,
					PriceHistory: nil,
				}

				continue
			}

			out <- GetPriceHistoryJob{
				Err:          nil,
				PriceHistory: receivedHistory,
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

	pHistory := sotah.PriceHistory{}
	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields(id)).Error("failed to fetch price history for item")

			return sotah.PriceHistory{}, job.Err
		}

		pHistory = pHistory.Merge(job.PriceHistory.Between(lowerBounds, upperBounds))
	}

	return pHistory, nil
}
