package stats

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type PersistRealmStatsInJob struct {
	Tuple            blizzardv2.RegionConnectedRealmTuple
	CurrentTimestamp sotah.UnixTimestamp
	Stats            sotah.MiniAuctionListStats
}

type PersistRealmStatsOutJob struct {
	Err   error
	Tuple blizzardv2.RegionConnectedRealmTuple
}

func (job PersistRealmStatsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (tBases TupleDatabases) PersistStats(in chan PersistRealmStatsInJob) error {
	out := make(chan PersistRealmStatsOutJob)

	worker := func() {
		for job := range in {
			tBase, err := tBases.GetTupleDatabase(job.Tuple)
			if err != nil {
				out <- PersistRealmStatsOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			err = tBase.Database.PersistStats(job.CurrentTimestamp, job.Stats)
			if err != nil {
				out <- PersistRealmStatsOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			out <- PersistRealmStatsOutJob{
				Err:   nil,
				Tuple: job.Tuple,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to persist stats")

			return job.Err
		}
	}

	return nil
}
