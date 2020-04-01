package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAuctionsJob struct {
	Err              error
	Tuple            RegionConnectedRealmTuple
	AuctionsResponse AuctionsResponse
}

func (job GetAuctionsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

type GetAuctionsOptions struct {
	Tuples         []RegionConnectedRealmTuple
	GetAuctionsURL func(RegionConnectedRealmTuple) (string, error)
}

func GetAuctions(opts GetAuctionsOptions) chan GetAuctionsJob {
	// starting up workers for gathering individual connected-realms
	in := make(chan RegionConnectedRealmTuple)
	out := make(chan GetAuctionsJob)
	worker := func() {
		for tuple := range in {
			getAuctionsUri, err := opts.GetAuctionsURL(tuple)
			if err != nil {
				out <- GetAuctionsJob{
					Err:              err,
					Tuple:            tuple,
					AuctionsResponse: AuctionsResponse{},
				}

				continue
			}

			auctionsResponse, _, err := NewAuctionsFromHTTP(getAuctionsUri)
			if err != nil {
				out <- GetAuctionsJob{
					Err:              err,
					Tuple:            tuple,
					AuctionsResponse: AuctionsResponse{},
				}

				continue
			}

			out <- GetAuctionsJob{
				Err:              nil,
				Tuple:            tuple,
				AuctionsResponse: auctionsResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	return out
}
