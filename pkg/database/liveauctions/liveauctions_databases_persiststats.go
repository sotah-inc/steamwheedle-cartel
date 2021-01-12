package liveauctions

import (
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type PersistRealmStatsJob struct {
	Err   error
	Tuple blizzardv2.RegionConnectedRealmTuple
}

func (job PersistRealmStatsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (ladBases Databases) PersistStats(tuples blizzardv2.RegionConnectedRealmTuples) error {
	in := make(chan blizzardv2.RegionConnectedRealmTuple)
	out := make(chan PersistRealmStatsJob)

	currentTimestamp := sotah.UnixTimestamp(time.Now().Unix())

	worker := func() {
		for tuple := range in {
			ladBase, err := ladBases.GetDatabase(tuple)
			if err != nil {
				out <- PersistRealmStatsJob{err, tuple}

				continue
			}

			err = ladBase.persistStats(currentTimestamp)
			if err != nil {
				out <- PersistRealmStatsJob{err, tuple}

				continue
			}

			out <- PersistRealmStatsJob{nil, tuple}
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
			logging.WithFields(job.ToLogrusFields()).Error("failed to persist auction-stats")

			return job.Err
		}
	}

	return nil
}
