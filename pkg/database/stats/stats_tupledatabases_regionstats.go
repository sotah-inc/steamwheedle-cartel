package stats

import (
	"errors"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type RegionStatsOutJob struct {
	Err          error
	Tuple        blizzardv2.RegionVersionConnectedRealmTuple
	AuctionStats sotah.AuctionStats
}

func (job RegionStatsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"game-version":    job.Tuple.Version,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (tBases TupleDatabases) RegionStats(
	regionName blizzardv2.RegionName,
) (sotah.AuctionStats, error) {
	// resolving databases
	regionTupleDatabases := tBases.GetTupleDatabasesByRegionName(regionName)
	if len(regionTupleDatabases) == 0 {
		return sotah.AuctionStats{}, errors.New("no stats databases for region")
	}

	in := make(chan TupleDatabase)
	out := make(chan RegionStatsOutJob)

	// spinning up workers for gathering stats
	worker := func() {
		for db := range in {
			auctionStats, err := db.Stats()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          db.tuple.RegionName,
					"connected-realm": db.tuple.ConnectedRealmId,
				}).Error("failed to get stats")

				out <- RegionStatsOutJob{
					Err:          err,
					Tuple:        db.tuple,
					AuctionStats: sotah.AuctionStats{},
				}

				continue
			}

			out <- RegionStatsOutJob{
				Err:          nil,
				Tuple:        db.tuple,
				AuctionStats: auctionStats,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		for _, db := range regionTupleDatabases {
			in <- db
		}

		close(in)
	}()

	// going over the results
	stats := sotah.AuctionStats{}
	for outJob := range out {
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("failed to gather stats")

			return sotah.AuctionStats{}, outJob.Err
		}

		stats = stats.Append(outJob.AuctionStats)

	}

	return stats, nil
}
