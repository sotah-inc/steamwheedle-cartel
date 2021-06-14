package disk

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (client Client) GetEncodedAuctionsByTuple(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) ([]byte, error) {
	cachedAuctionsFilepath, err := client.resolveAuctionsFilepath(tuple)
	if err != nil {
		return nil, err
	}

	return util.ReadFile(cachedAuctionsFilepath)
}

type getEncodedAuctionsByTuplesJob struct {
	err             error
	tuple           blizzardv2.RegionVersionConnectedRealmTuple
	encodedAuctions []byte
}

func (job getEncodedAuctionsByTuplesJob) Tuple() blizzardv2.RegionVersionConnectedRealmTuple {
	return job.tuple
}
func (job getEncodedAuctionsByTuplesJob) EncodedAuctions() []byte { return job.encodedAuctions }
func (job getEncodedAuctionsByTuplesJob) Err() error              { return job.err }
func (job getEncodedAuctionsByTuplesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.err.Error(),
		"region":          job.tuple.RegionName,
		"connected-realm": job.tuple.ConnectedRealmId,
	}
}

func (client Client) GetEncodedAuctionsByTuples(
	tuples blizzardv2.RegionVersionConnectedRealmTuples,
) chan BaseLake.GetEncodedAuctionsByTuplesJob {
	// establishing channels
	out := make(chan BaseLake.GetEncodedAuctionsByTuplesJob)
	in := make(chan blizzardv2.RegionVersionConnectedRealmTuple)

	// spinning up the workers for fetching auctions
	worker := func() {
		for tuple := range in {
			gzipEncoded, err := client.GetEncodedAuctionsByTuple(tuple)
			if err != nil {
				out <- getEncodedAuctionsByTuplesJob{
					err:             err,
					tuple:           tuple,
					encodedAuctions: nil,
				}

				continue
			}

			out <- getEncodedAuctionsByTuplesJob{err, tuple, gzipEncoded}
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
				"connected-realm": tuple.ConnectedRealmId,
			}).Debug("queueing up tuple for fetching")

			in <- tuple
		}

		close(in)
	}()

	return out
}
