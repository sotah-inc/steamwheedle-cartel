package disk

import (
	"os"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (client Client) getTupleStats(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (sotah.MiniAuctionListGeneralStats, error) {
	cachedAuctionsFilepath, err := client.resolveAuctionsFilepath(tuple)
	if err != nil {
		return sotah.MiniAuctionListGeneralStats{}, err
	}

	if _, err := os.Stat(cachedAuctionsFilepath); err != nil {
		if os.IsNotExist(err) {
			return sotah.MiniAuctionListGeneralStats{}, nil
		}

		return sotah.MiniAuctionListGeneralStats{}, err
	}

	gzipEncoded, err := util.ReadFile(cachedAuctionsFilepath)
	if err != nil {
		return sotah.MiniAuctionListGeneralStats{}, err
	}

	maList, err := sotah.NewMiniAuctionListFromGzipped(gzipEncoded)
	if err != nil {
		return sotah.MiniAuctionListGeneralStats{}, err
	}

	return sotah.NewMiniAuctionListStatsFromMiniAuctionList(maList).MiniAuctionListGeneralStats, nil
}

type getEncodedRegionStatsJob struct {
	err   error
	tuple blizzardv2.RegionVersionConnectedRealmTuple
	stats sotah.MiniAuctionListGeneralStats
}

func (job getEncodedRegionStatsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.err.Error(),
		"region":          job.tuple.RegionName,
		"game-version":    job.tuple.Version,
		"connected-realm": job.tuple.ConnectedRealmId,
	}
}

func (client Client) GetEncodedRegionStats(
	tuple blizzardv2.RegionVersionTuple,
	ids []blizzardv2.ConnectedRealmId,
) ([]byte, error) {
	in := make(chan blizzardv2.ConnectedRealmId)
	out := make(chan getEncodedRegionStatsJob)

	// spinning up the workers for fetching auctions
	worker := func() {
		for id := range in {
			stats, err := client.getTupleStats(blizzardv2.RegionVersionConnectedRealmTuple{
				RegionVersionTuple: tuple,
				ConnectedRealmId:   id,
			})
			if err != nil {
				out <- getEncodedRegionStatsJob{
					err: err,
					tuple: blizzardv2.RegionVersionConnectedRealmTuple{
						RegionVersionTuple: tuple,
						ConnectedRealmId:   id,
					},
					stats: sotah.MiniAuctionListGeneralStats{},
				}

				continue
			}

			if stats.TotalAuctions == 0 {
				logging.WithFields(logrus.Fields{
					"region":          tuple.RegionName,
					"game-version":    tuple.Version,
					"connected-realm": id,
				}).Info("no stats were found for region/connected-realm")

				continue
			}

			out <- getEncodedRegionStatsJob{
				err: nil,
				tuple: blizzardv2.RegionVersionConnectedRealmTuple{
					RegionVersionTuple: tuple,
					ConnectedRealmId:   id,
				},
				stats: stats,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing up the tuples
	go func() {
		for _, id := range ids {
			logging.WithFields(logrus.Fields{
				"region":          tuple.RegionName,
				"game-version":    tuple.Version,
				"connected-realm": id,
			}).Debug("queueing up tuple for fetching")

			in <- id
		}

		close(in)
	}()

	// going over the results
	stats := sotah.MiniAuctionListGeneralStats{}
	for job := range out {
		if job.err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to resolve stats")

			return []byte{}, job.err
		}

		stats = stats.Add(job.stats)
	}

	return stats.EncodeForStorage()
}
