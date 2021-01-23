package disk

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (client Client) getTupleStats(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (sotah.MiniAuctionListGeneralStats, error) {
	cachedAuctionsFilepath, err := client.resolveAuctionsFilepath(tuple)
	if err != nil {
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
	err              error
	name             blizzardv2.RegionName
	connectedRealmId blizzardv2.ConnectedRealmId
	stats            sotah.MiniAuctionListGeneralStats
}

func (job getEncodedRegionStatsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.err.Error(),
		"region":          job.name,
		"connected-realm": job.connectedRealmId,
	}
}

func (client Client) GetEncodedRegionStats(
	name blizzardv2.RegionName,
	ids []blizzardv2.ConnectedRealmId,
) ([]byte, error) {
	in := make(chan blizzardv2.ConnectedRealmId)
	out := make(chan getEncodedRegionStatsJob)

	// spinning up the workers for fetching auctions
	worker := func() {
		for id := range in {
			stats, err := client.getTupleStats(blizzardv2.RegionConnectedRealmTuple{
				RegionName:       name,
				ConnectedRealmId: id,
			})
			if err != nil {
				out <- getEncodedRegionStatsJob{
					err:              err,
					name:             name,
					connectedRealmId: id,
					stats:            sotah.MiniAuctionListGeneralStats{},
				}

				continue
			}

			out <- getEncodedRegionStatsJob{
				err:              nil,
				name:             name,
				connectedRealmId: id,
				stats:            stats,
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
				"region":          name,
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
