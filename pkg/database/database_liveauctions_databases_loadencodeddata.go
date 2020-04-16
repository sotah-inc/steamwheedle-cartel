package database

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type LiveAuctionsLoadEncodedDataInJob struct {
	Tuple       blizzardv2.RegionConnectedRealmTuple
	EncodedData []byte
}

type LiveAuctionsLoadEncodedDataOutJob struct {
	Err   error
	Tuple blizzardv2.RegionConnectedRealmTuple
}

func (job LiveAuctionsLoadEncodedDataOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (ladBases LiveAuctionsDatabases) LoadEncodedData(
	in chan LiveAuctionsLoadEncodedDataInJob,
) chan LiveAuctionsLoadEncodedDataOutJob {
	// establishing channels
	out := make(chan LiveAuctionsLoadEncodedDataOutJob)

	// spinning up workers for receiving encoded-data and persisting it
	worker := func() {
		for job := range in {
			// resolving the live-auctions database and gathering current Stats
			ladBase, err := ladBases.GetDatabase(job.Tuple)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          job.Tuple.RegionName,
					"connected-realm": job.Tuple.ConnectedRealmId,
				}).Error("failed to find database by tuple")

				out <- LiveAuctionsLoadEncodedDataOutJob{
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

				out <- LiveAuctionsLoadEncodedDataOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			out <- LiveAuctionsLoadEncodedDataOutJob{
				Err:   nil,
				Tuple: job.Tuple,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	return out
}
