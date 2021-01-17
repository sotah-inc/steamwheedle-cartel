package stats

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type PersistRealmStatsInJob struct {
	Tuple        blizzardv2.LoadConnectedRealmTuple
	EncodedStats []byte
}

type PersistRealmStatsOutJob struct {
	Err   error
	Tuple blizzardv2.LoadConnectedRealmTuple
}

func (job PersistRealmStatsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (tBases TupleDatabases) PersistEncodedStats(in chan PersistRealmStatsInJob) chan PersistRealmStatsOutJob {
	out := make(chan PersistRealmStatsOutJob)

	worker := func() {
		for job := range in {
			tBase, err := tBases.GetTupleDatabase(job.Tuple.RegionConnectedRealmTuple)
			if err != nil {
				out <- PersistRealmStatsOutJob{
					Err:   err,
					Tuple: job.Tuple,
				}

				continue
			}

			err = tBase.Database.PersistEncodedStats(
				sotah.UnixTimestamp(job.Tuple.LastModified.Unix()),
				job.EncodedStats,
			)
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

	return out
}
