package database

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewLiveAuctionsDatabases(
	dirPath string,
	tuples []blizzardv2.RegionConnectedRealmTuple,
) (LiveAuctionsDatabases, error) {
	ladBases := LiveAuctionsDatabases{}

	for _, tuple := range tuples {
		shards := func() LiveAuctionsDatabaseShards {
			out, ok := ladBases[tuple.RegionName]
			if !ok {
				return LiveAuctionsDatabaseShards{}
			}

			return out
		}()

		var err error
		shards[tuple.ConnectedRealmId], err = newLiveAuctionsDatabase(dirPath, tuple)
		if err != nil {
			return LiveAuctionsDatabases{}, err
		}

		ladBases[tuple.RegionName] = shards
	}

	return ladBases, nil
}

type LiveAuctionsDatabases map[blizzardv2.RegionName]LiveAuctionsDatabaseShards

func (ladBases LiveAuctionsDatabases) GetDatabase(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (LiveAuctionsDatabase, error) {
	shards, ok := ladBases[tuple.RegionName]
	if !ok {
		return LiveAuctionsDatabase{}, fmt.Errorf("shard not found for region %s", tuple.RegionName)
	}

	db, ok := shards[tuple.ConnectedRealmId]
	if !ok {
		return LiveAuctionsDatabase{}, fmt.Errorf("db not found for connected-realm %d", tuple.ConnectedRealmId)
	}

	return db, nil
}

type AuctionStatsWithTuplesOutJob struct {
	Err          error
	Tuple        blizzardv2.RegionConnectedRealmTuple
	AuctionStats sotah.AuctionStats
}

func (job AuctionStatsWithTuplesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

func (ladBases LiveAuctionsDatabases) AuctionStatsWithTuples(
	tuples blizzardv2.RegionConnectedRealmTuples,
) (sotah.AuctionStats, error) {
	in := make(chan blizzardv2.RegionConnectedRealmTuple)
	out := make(chan AuctionStatsWithTuplesOutJob)

	// spinning up workers for gathering stats
	worker := func() {
		for tuple := range in {
			// resolving the live-auctions database and gathering current Stats
			ladBase, err := ladBases.GetDatabase(tuple)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          tuple.RegionName,
					"connected-realm": tuple.ConnectedRealmId,
				}).Error("failed to find database by tuple")

				out <- AuctionStatsWithTuplesOutJob{
					Err:   err,
					Tuple: tuple,
				}

				continue
			}

			auctionStats, err := ladBase.AuctionStats()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":           err.Error(),
					"region":          tuple.RegionName,
					"connected-realm": tuple.ConnectedRealmId,
				}).Error("failed to get auction-stats")

				out <- AuctionStatsWithTuplesOutJob{
					Err:   err,
					Tuple: tuple,
				}

				continue
			}

			out <- AuctionStatsWithTuplesOutJob{
				Err:          nil,
				Tuple:        tuple,
				AuctionStats: auctionStats,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing it up
	go func() {
		for _, tuple := range tuples {
			in <- tuple
		}

		close(in)
	}()

	// going over the results
	results := sotah.AuctionStats{}
	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to fetch auction-stats")

			return sotah.AuctionStats{}, job.Err
		}

		results = results.Append(job.AuctionStats)
	}

	return results, nil
}

type LiveAuctionsDatabaseShards map[blizzardv2.ConnectedRealmId]LiveAuctionsDatabase
