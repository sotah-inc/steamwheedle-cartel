package disk

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (client Client) GetEncodedRecipePricesByTuple(
	mRecipes sotah.MiniRecipes,
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (map[blizzardv2.RecipeId][]byte, error) {
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

	out := map[blizzardv2.RecipeId][]byte{}
	for id, prices := range sotah.NewRecipePricesMap(
		mRecipes,
		sotah.NewItemPricesFromMiniAuctionList(maList),
	) {
		out[id], err = prices.EncodeForStorage()
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

type getEncodedRecipePricesByTuplesJob struct {
	err                 error
	tuple               blizzardv2.LoadConnectedRealmTuple
	encodedRecipePrices map[blizzardv2.RecipeId][]byte
}

func (job getEncodedRecipePricesByTuplesJob) Tuple() blizzardv2.LoadConnectedRealmTuple {
	return job.tuple
}
func (job getEncodedRecipePricesByTuplesJob) EncodedRecipePrices() map[blizzardv2.RecipeId][]byte {
	return job.encodedRecipePrices
}
func (job getEncodedRecipePricesByTuplesJob) Err() error { return job.err }
func (job getEncodedRecipePricesByTuplesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.err.Error(),
		"region":          job.tuple.RegionName,
		"game-version":    job.tuple.Version,
		"connected-realm": job.tuple.ConnectedRealmId,
	}
}

func (client Client) GetEncodedRecipePricesByTuples(
	mRecipes sotah.MiniRecipes,
	tuples blizzardv2.LoadConnectedRealmTuples,
) chan BaseLake.GetEncodedRecipePricesByTuplesJob {
	// establishing channels
	out := make(chan BaseLake.GetEncodedRecipePricesByTuplesJob)
	in := make(chan blizzardv2.LoadConnectedRealmTuple)

	// spinning up the workers for fetching auctions
	worker := func() {
		for tuple := range in {
			gzipEncoded, err := client.GetEncodedRecipePricesByTuple(
				mRecipes,
				tuple.RegionVersionConnectedRealmTuple,
			)
			if err != nil {
				out <- getEncodedRecipePricesByTuplesJob{
					err:                 err,
					tuple:               tuple,
					encodedRecipePrices: nil,
				}

				continue
			}

			out <- getEncodedRecipePricesByTuplesJob{err, tuple, gzipEncoded}
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
