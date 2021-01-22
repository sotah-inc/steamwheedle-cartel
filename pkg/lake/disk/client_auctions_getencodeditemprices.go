package disk

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (client Client) GetEncodedPricelistHistoryByTuple(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (map[blizzardv2.ItemId][]byte, error) {
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

	out := map[blizzardv2.ItemId][]byte{}
	for id, prices := range sotah.NewItemPricesFromMiniAuctionList(maList) {
		out[id], err = prices.EncodeForStorage()
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

type getEncodedPricelistHistoryByTuplesJob struct {
	err                     error
	tuple                   blizzardv2.LoadConnectedRealmTuple
	encodedPricelistHistory map[blizzardv2.ItemId][]byte
}

func (job getEncodedPricelistHistoryByTuplesJob) Tuple() blizzardv2.LoadConnectedRealmTuple {
	return job.tuple
}
func (
	job getEncodedPricelistHistoryByTuplesJob,
) EncodedPricelistHistory() map[blizzardv2.ItemId][]byte {
	return job.encodedPricelistHistory
}
func (job getEncodedPricelistHistoryByTuplesJob) Err() error { return job.err }
func (job getEncodedPricelistHistoryByTuplesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.err.Error(),
		"region":          job.tuple.RegionName,
		"connected-realm": job.tuple.ConnectedRealmId,
	}
}

func (client Client) GetEncodedItemPricesByTuples(
	tuples blizzardv2.LoadConnectedRealmTuples,
) chan BaseLake.GetEncodedItemPricesByTuplesJob {
	// establishing channels
	out := make(chan BaseLake.GetEncodedItemPricesByTuplesJob)
	in := make(chan blizzardv2.LoadConnectedRealmTuple)

	// spinning up the workers for fetching auctions
	worker := func() {
		for tuple := range in {
			gzipEncoded, err := client.GetEncodedPricelistHistoryByTuple(tuple.RegionConnectedRealmTuple)
			if err != nil {
				out <- getEncodedPricelistHistoryByTuplesJob{
					err:                     err,
					tuple:                   tuple,
					encodedPricelistHistory: nil,
				}

				continue
			}

			out <- getEncodedPricelistHistoryByTuplesJob{err, tuple, gzipEncoded}
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
