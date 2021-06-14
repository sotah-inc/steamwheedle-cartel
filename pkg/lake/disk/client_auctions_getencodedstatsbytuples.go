package disk

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (client Client) GetEncodedStatsByTuple(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) ([]byte, error) {
	cachedAuctionsFilepath, err := client.resolveAuctionsFilepath(tuple)
	if err != nil {
		return nil, err
	}

	gzipEncoded, err := util.ReadFile(cachedAuctionsFilepath)
	if err != nil {
		return nil, err
	}

	maList, err := sotah.NewMiniAuctionListFromGzipped(gzipEncoded)
	if err != nil {
		return nil, err
	}

	encodedStats, err := sotah.NewMiniAuctionListStatsFromMiniAuctionList(maList).EncodeForStorage()
	if err != nil {
		return nil, err
	}

	return encodedStats, nil
}

type getEncodedStatsByTuplesJob struct {
	err          error
	tuple        blizzardv2.LoadConnectedRealmTuple
	encodedStats []byte
}

func (job getEncodedStatsByTuplesJob) Tuple() blizzardv2.LoadConnectedRealmTuple {
	return job.tuple
}
func (job getEncodedStatsByTuplesJob) EncodedStats() []byte { return job.encodedStats }
func (job getEncodedStatsByTuplesJob) Err() error           { return job.err }
func (job getEncodedStatsByTuplesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.err.Error(),
		"region":          job.tuple.RegionName,
		"game-version":    job.tuple.Version,
		"connected-realm": job.tuple.ConnectedRealmId,
	}
}

func (client Client) GetEncodedStatsByTuples(
	tuples blizzardv2.LoadConnectedRealmTuples,
) chan BaseLake.GetEncodedStatsByTuplesJob {
	// establishing channels
	out := make(chan BaseLake.GetEncodedStatsByTuplesJob)
	in := make(chan blizzardv2.LoadConnectedRealmTuple)

	// spinning up the workers for fetching auctions
	worker := func() {
		for tuple := range in {
			jsonEncoded, err := client.GetEncodedStatsByTuple(tuple.RegionVersionConnectedRealmTuple)
			if err != nil {
				out <- getEncodedStatsByTuplesJob{
					err:          err,
					tuple:        tuple,
					encodedStats: nil,
				}

				continue
			}

			out <- getEncodedStatsByTuplesJob{
				err:          nil,
				tuple:        tuple,
				encodedStats: jsonEncoded,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the tuples
	go func() {
		for _, tuple := range tuples {
			logging.WithFields(logrus.Fields{
				"region":          tuple.RegionName,
				"game-version":    tuple.Version,
				"connected-realm": tuple.ConnectedRealmId,
			}).Debug("queueing up tuple for fetching")

			in <- tuple
		}

		close(in)
	}()

	return out
}
