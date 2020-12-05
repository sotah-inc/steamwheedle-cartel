package pricelisthistory

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type LoadEncodedRecipePricesInJob struct {
	Tuple       blizzardv2.LoadConnectedRealmTuple
	EncodedData map[blizzardv2.RecipeId][]byte
}

type LoadEncodedRecipePricesOutJob struct {
	Err        error
	Tuple      blizzardv2.LoadConnectedRealmTuple
	ReceivedAt time.Time
}

func (job LoadEncodedRecipePricesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (phdBases *Databases) LoadEncodedRecipePrices(
	in chan LoadEncodedRecipePricesInJob,
) chan LoadEncodedRecipePricesOutJob {
	// establishing channels
	out := make(chan LoadEncodedRecipePricesOutJob)

	// spinning up workers for receiving encoded-data and persisting it
	worker := func() {
		for job := range in {
			// resolving the live-auctions database and gathering current Stats
			phdBase, err := phdBases.resolveDatabase(job.Tuple)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          job.Tuple.RegionName,
					"connected-realm": job.Tuple.ConnectedRealmId,
				}).Error("failed to find database by tuple")

				out <- LoadEncodedRecipePricesOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			if err := phdBase.persistEncodedRecipePrices(job.EncodedData); err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          job.Tuple.RegionName,
					"connected-realm": job.Tuple.ConnectedRealmId,
				}).Error("failed to persist encoded-data")

				out <- LoadEncodedRecipePricesOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			out <- LoadEncodedRecipePricesOutJob{
				Err:        nil,
				Tuple:      job.Tuple,
				ReceivedAt: time.Now(),
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	return out
}
