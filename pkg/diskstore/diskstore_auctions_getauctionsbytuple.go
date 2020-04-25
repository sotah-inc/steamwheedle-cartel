package diskstore

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (ds DiskStore) GetEncodedAuctionsByTuple(
	tuple blizzardv2.RegionConnectedRealmTuple,
) ([]byte, error) {
	cachedAuctionsFilepath, err := ds.resolveAuctionsFilepath(tuple)
	if err != nil {
		return nil, err
	}

	return util.ReadFile(cachedAuctionsFilepath)
}

type GetEncodedAuctionsByTuplesJob struct {
	Err             error
	Tuple           blizzardv2.RegionConnectedRealmTuple
	EncodedAuctions []byte
}

func (job GetEncodedAuctionsByTuplesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{}
}

func (ds DiskStore) GetEncodedAuctionsByTuples(
	tuples []blizzardv2.RegionConnectedRealmTuple,
) chan GetEncodedAuctionsByTuplesJob {
	// establishing channels
	out := make(chan GetEncodedAuctionsByTuplesJob)
	in := make(chan blizzardv2.RegionConnectedRealmTuple)

	// spinning up the workers for fetching auctions
	worker := func() {
		for tuple := range in {
			gzipEncoded, err := ds.GetEncodedAuctionsByTuple(tuple)
			if err != nil {
				out <- GetEncodedAuctionsByTuplesJob{
					Err:             err,
					Tuple:           tuple,
					EncodedAuctions: nil,
				}

				continue
			}

			out <- GetEncodedAuctionsByTuplesJob{err, tuple, gzipEncoded}
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
