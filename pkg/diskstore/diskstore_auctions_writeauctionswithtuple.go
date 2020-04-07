package diskstore

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (ds DiskStore) WriteAuctionsWithTuple(tuple blizzardv2.RegionConnectedRealmTuple, auctions blizzardv2.Auctions) error {
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
	Tuple    blizzardv2.RegionConnectedRealmTuple
	Auctions blizzardv2.Auctions
}

type WriteAuctionsWithTuplesOutJob struct {
	Err   error
	Tuple blizzardv2.RegionConnectedRealmTuple
}

func (job WriteAuctionsWithTuplesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
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
				"region":          job.Tuple.RegionName,
				"connected-realm": job.Tuple.ConnectedRealmId,
			}).Debug("queueing up job for writing auctions")

			in <- job
		}
	}()

	return out
}
