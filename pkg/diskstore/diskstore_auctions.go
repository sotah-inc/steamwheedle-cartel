package diskstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (ds DiskStore) resolveAuctionsFilepath(tuple sotah.RegionRealmTuple) (string, error) {
	if len(ds.CacheDir) == 0 {
		return "", errors.New("cache dir cannot be blank")
	}

	return fmt.Sprintf("%s/auctions/%s/%s.json.gz", ds.CacheDir, tuple.RegionName, tuple.RealmSlug), nil
}

func (ds DiskStore) WriteAuctionsWithTuple(tuple sotah.RegionRealmTuple, auctions blizzardv2.Auctions) error {
	dest, err := ds.resolveAuctionsFilepath(tuple)
	if err != nil {
		return err
	}

	jsonEncoded, err := json.Marshal(auctions)
	if err != nil {
		return err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return err
	}

	return util.WriteFile(dest, gzipEncoded)
}

type WriteAuctionsWithTuplesInJob struct {
	Tuple    sotah.RegionRealmTuple
	Auctions blizzardv2.Auctions
}

type WriteAuctionsWithTuplesOutJob struct {
	Err   error
	Tuple sotah.RegionRealmTuple
}

func (job WriteAuctionsWithTuplesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"region": job.Tuple.RegionName,
		"realm":  job.Tuple.RealmSlug,
	}
}

func (ds DiskStore) WriteAuctionsWithTuples(in chan WriteAuctionsWithTuplesInJob) chan WriteAuctionsWithTuplesOutJob {
	// establishing channels
	out := make(chan WriteAuctionsWithTuplesOutJob)

	// spinning up the workers for writing
	worker := func() {
		for job := range in {
			if err := ds.WriteAuctionsWithTuple(job.Tuple, job.Auctions); err != nil {
				out <- WriteAuctionsWithTuplesOutJob{err, job.Tuple}

				continue
			}

			out <- WriteAuctionsWithTuplesOutJob{nil, job.Tuple}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing up the jobs
	go func() {
		for job := range in {
			logging.WithFields(logrus.Fields{
				"region": job.Tuple.RegionName,
				"realm":  job.Tuple.RealmSlug,
			}).Debug("queueing up job for writing auctions")

			in <- job
		}
	}()

	return out
}

func (ds DiskStore) GetAuctionsByTuple(tuple sotah.RegionRealmTuple) (blizzardv2.Auctions, time.Time, error) {
	// resolving the cached auctions filepath
	cachedAuctionsFilepath, err := ds.resolveAuctionsFilepath(tuple)
	if err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	// optionally skipping non-exist auctions file
	cachedAuctionsStat, err := os.Stat(cachedAuctionsFilepath)
	if err != nil {
		if !os.IsNotExist(err) {
			return blizzardv2.Auctions{}, time.Time{}, err
		}

		return blizzardv2.Auctions{}, time.Time{}, nil
	}

	gzipEncoded, err := util.ReadFile(cachedAuctionsFilepath)
	if err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	var auctions blizzardv2.Auctions
	if err := json.Unmarshal(jsonEncoded, &auctions); err != nil {
		return blizzardv2.Auctions{}, time.Time{}, err
	}

	return auctions, cachedAuctionsStat.ModTime(), nil
}

type GetAuctionsByTuplesJob struct {
	Err          error
	Tuple        sotah.RegionRealmTuple
	Auctions     blizzardv2.Auctions
	LastModified time.Time
}

func (job GetAuctionsByTuplesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{}
}

func (ds DiskStore) GetAuctionsByTuples(tuples sotah.RegionRealmTuples) chan GetAuctionsByTuplesJob {
	// establishing channels
	out := make(chan GetAuctionsByTuplesJob)
	in := make(chan sotah.RegionRealmTuple)

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
					"region": tuple.RegionName,
					"realm":  tuple.RealmSlug,
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
				"region": tuple.RegionName,
				"realm":  tuple.RealmSlug,
			}).Debug("queueing up tuple for fetching")

			in <- tuple
		}

		close(in)
	}()

	return out
}
