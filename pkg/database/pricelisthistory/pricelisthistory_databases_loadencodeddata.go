package pricelisthistory

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type LoadEncodedDataInJob struct {
	Tuple       blizzardv2.LoadConnectedRealmTuple
	EncodedData map[blizzardv2.ItemId][]byte
}

type LoadEncodedDataOutJob struct {
	Err        error
	Tuple      blizzardv2.LoadConnectedRealmTuple
	ReceivedAt time.Time
}

func (job LoadEncodedDataOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (phdBases *Databases) LoadEncodedData(
	in chan LoadEncodedDataInJob,
) chan LoadEncodedDataOutJob {
	// establishing channels
	out := make(chan LoadEncodedDataOutJob)

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

				out <- LoadEncodedDataOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			if err := phdBase.persistEncodedData(job.EncodedData); err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          job.Tuple.RegionName,
					"connected-realm": job.Tuple.ConnectedRealmId,
				}).Error("failed to persist encoded-data")

				out <- LoadEncodedDataOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			out <- LoadEncodedDataOutJob{
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
