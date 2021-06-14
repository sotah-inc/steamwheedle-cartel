package disk

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (client Client) WriteAuctionsWithTuple(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	auctions sotah.MiniAuctionList,
) error {
	dest, err := client.resolveAuctionsFilepath(tuple)
	if err != nil {
		return err
	}

	gzipEncoded, err := auctions.EncodeForStorage()
	if err != nil {
		return err
	}

	return util.WriteFile(dest, gzipEncoded)
}

func (client Client) NewWriteAuctionsWithTuplesInJob(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	auctions sotah.MiniAuctionList,
) BaseLake.WriteAuctionsWithTuplesInJob {
	return WriteAuctionsWithTuplesInJob{
		tuple:    tuple,
		auctions: auctions,
	}
}

type WriteAuctionsWithTuplesInJob struct {
	tuple    blizzardv2.RegionVersionConnectedRealmTuple
	auctions sotah.MiniAuctionList
}

func (w WriteAuctionsWithTuplesInJob) Tuple() blizzardv2.RegionVersionConnectedRealmTuple {
	return w.tuple
}
func (w WriteAuctionsWithTuplesInJob) Auctions() sotah.MiniAuctionList { return w.auctions }

type WriteAuctionsWithTuplesOutJob struct {
	err   error
	tuple blizzardv2.RegionVersionConnectedRealmTuple
}

func (job WriteAuctionsWithTuplesOutJob) Err() error { return job.err }
func (job WriteAuctionsWithTuplesOutJob) Tuple() blizzardv2.RegionVersionConnectedRealmTuple {
	return job.tuple
}
func (job WriteAuctionsWithTuplesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err().Error(),
		"region":          job.Tuple().RegionName,
		"game-version":    job.tuple.Version,
		"connected-realm": job.Tuple().ConnectedRealmId,
	}
}

func (client Client) WriteAuctionsWithTuples(
	in chan BaseLake.WriteAuctionsWithTuplesInJob,
) chan BaseLake.WriteAuctionsWithTuplesOutJob {
	// establishing channels
	out := make(chan BaseLake.WriteAuctionsWithTuplesOutJob)

	// spinning up the workers for writing
	worker := func() {
		for job := range in {
			if err := client.WriteAuctionsWithTuple(job.Tuple(), job.Auctions()); err != nil {
				out <- WriteAuctionsWithTuplesOutJob{err, job.Tuple()}

				continue
			}

			out <- WriteAuctionsWithTuplesOutJob{nil, job.Tuple()}
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
				"region":          job.Tuple().RegionName,
				"game-version":    job.Tuple().Version,
				"connected-realm": job.Tuple().ConnectedRealmId,
			}).Debug("queueing up job for writing auctions")

			in <- job
		}
	}()

	return out
}
