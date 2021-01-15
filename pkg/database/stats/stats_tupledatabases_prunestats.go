package stats

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type PruneStatsJob struct {
	Err   error
	Tuple blizzardv2.RegionConnectedRealmTuple
}

func (job PruneStatsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (tBases TupleDatabases) PruneStats(
	tuples blizzardv2.RegionConnectedRealmTuples,
	retentionLimit sotah.UnixTimestamp,
) error {
	in := make(chan blizzardv2.RegionConnectedRealmTuple)
	out := make(chan PruneStatsJob)

	worker := func() {
		for tuple := range in {
			ladBase, err := tBases.GetDatabase(tuple)
			if err != nil {
				out <- PruneStatsJob{err, tuple}

				continue
			}

			err = ladBase.pruneStats(retentionLimit)
			if err != nil {
				out <- PruneStatsJob{err, tuple}

				continue
			}

			out <- PruneStatsJob{nil, tuple}
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
			logging.WithFields(job.ToLogrusFields()).Error("failed to prune auction-stats")

			return job.Err
		}
	}

	return nil
}
