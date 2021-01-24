package stats

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type PruneRealmStatsJob struct {
	Err   error
	Tuple blizzardv2.RegionConnectedRealmTuple
}

func (job PruneRealmStatsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (tBases TupleDatabases) PruneRealmStats(
	tuples blizzardv2.RegionConnectedRealmTuples,
	retentionLimit sotah.UnixTimestamp,
) error {
	in := make(chan blizzardv2.RegionConnectedRealmTuple)
	out := make(chan PruneRealmStatsJob)

	worker := func() {
		for tuple := range in {
			ladBase, err := tBases.GetTupleDatabase(tuple)
			if err != nil {
				out <- PruneRealmStatsJob{err, tuple}

				continue
			}

			err = ladBase.PruneStats(retentionLimit)
			if err != nil {
				out <- PruneRealmStatsJob{err, tuple}

				continue
			}

			out <- PruneRealmStatsJob{nil, tuple}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	go func() {
		for _, tuple := range tuples {
			in <- tuple
		}

		close(in)
	}()

	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to prune stats")

			return job.Err
		}
	}

	return nil
}
