package diskstore

import (
	"encoding/json"
	"os"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (ds DiskStore) GetAuctionsByTuple(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (sotah.MiniAuctionList, time.Time, error) {
	// resolving the cached auctions filepath
	cachedAuctionsFilepath, err := ds.resolveAuctionsFilepath(tuple)
	if err != nil {
		return sotah.MiniAuctionList{}, time.Time{}, err
	}

	// optionally skipping non-exist auctions file
	cachedAuctionsStat, err := os.Stat(cachedAuctionsFilepath)
	if err != nil {
		if !os.IsNotExist(err) {
			return sotah.MiniAuctionList{}, time.Time{}, err
		}

		return sotah.MiniAuctionList{}, time.Time{}, nil
	}

	gzipEncoded, err := util.ReadFile(cachedAuctionsFilepath)
	if err != nil {
		return sotah.MiniAuctionList{}, time.Time{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return sotah.MiniAuctionList{}, time.Time{}, err
	}

	var out sotah.MiniAuctionList
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return sotah.MiniAuctionList{}, time.Time{}, err
	}

	return out, cachedAuctionsStat.ModTime(), nil
}

type GetAuctionsByTuplesJob struct {
	Err          error
	Tuple        blizzardv2.RegionConnectedRealmTuple
	Auctions     sotah.MiniAuctionList
	LastModified time.Time
}

func (job GetAuctionsByTuplesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{}
}

func (ds DiskStore) GetAuctionsByTuples(tuples []blizzardv2.RegionConnectedRealmTuple) chan GetAuctionsByTuplesJob {
	// establishing channels
	out := make(chan GetAuctionsByTuplesJob)
	in := make(chan blizzardv2.RegionConnectedRealmTuple)

	// spinning up the workers for fetching auctions
	worker := func() {
		for tuple := range in {
			aucs, lastModified, err := ds.GetAuctionsByTuple(tuple)
			if err != nil {
				out <- GetAuctionsByTuplesJob{
					Err:          err,
					Tuple:        tuple,
					Auctions:     nil,
					LastModified: time.Time{},
				}

				continue
			}

			if lastModified.IsZero() {
				logging.WithFields(logrus.Fields{
					"region":          tuple.RegionName,
					"connected-realm": tuple.ConnectedRealmId,
				}).Error("last-modified was blank when fetching auctions from filecache")

				continue
			}

			out <- GetAuctionsByTuplesJob{err, tuple, aucs, lastModified}
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
