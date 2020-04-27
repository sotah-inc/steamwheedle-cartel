package database

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type PricelistHistoryLoadEncodedDataInJob struct {
	Tuple       blizzardv2.LoadConnectedRealmTuple
	EncodedData map[blizzardv2.ItemId][]byte
}

type PricelistHistoryLoadEncodedDataOutJob struct {
	Err        error
	Tuple      blizzardv2.LoadConnectedRealmTuple
	ReceivedAt time.Time
}

func (job PricelistHistoryLoadEncodedDataOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (phdBases *PricelistHistoryDatabases) LoadEncodedData(
	in chan PricelistHistoryLoadEncodedDataInJob,
) chan PricelistHistoryLoadEncodedDataOutJob {
	// establishing channels
	out := make(chan PricelistHistoryLoadEncodedDataOutJob)

	// spinning up workers for receiving encoded-data and persisting it
	worker := func() {
		for job := range in {
			// resolving the live-auctions database and gathering current Stats
			ladBase, err := phdBases.GetDatabase(job.Tuple)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          job.Tuple.RegionName,
					"connected-realm": job.Tuple.ConnectedRealmId,
				}).Error("failed to find database by tuple")

				out <- PricelistHistoryLoadEncodedDataOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			if err := ladBase.persistEncodedData(job.EncodedData); err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          job.Tuple.RegionName,
					"connected-realm": job.Tuple.ConnectedRealmId,
				}).Error("failed to persist encoded-data")

				out <- PricelistHistoryLoadEncodedDataOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			out <- PricelistHistoryLoadEncodedDataOutJob{
				Err:        nil,
				Tuple:      job.Tuple,
				ReceivedAt: time.Now(),
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	return out
}
