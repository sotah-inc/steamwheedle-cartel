package blizzardv2

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAuctionsJob struct {
	Err              error
	Tuple            LoadConnectedRealmTuple
	AuctionsResponse AuctionsResponse
}

func (job GetAuctionsJob) IsNew() bool {
	return !job.Tuple.LastModified.IsZero()
}

func (job GetAuctionsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":           job.Err.Error(),
		"region":          job.Tuple.RegionName,
		"connected-realm": job.Tuple.ConnectedRealmId,
	}
}

type GetAuctionsOptions struct {
	Tuples         []DownloadConnectedRealmTuple
	GetAuctionsURL func(DownloadConnectedRealmTuple) (string, error)
}

func GetAuctions(opts GetAuctionsOptions) chan GetAuctionsJob {
	// starting up workers for gathering individual connected-realms
	in := make(chan DownloadConnectedRealmTuple)
	out := make(chan GetAuctionsJob)
	worker := func() {
		for tuple := range in {
			getAuctionsUri, err := opts.GetAuctionsURL(tuple)
			if err != nil {
				out <- GetAuctionsJob{
					Err: err,
					Tuple: LoadConnectedRealmTuple{
						RegionVersionConnectedRealmTuple: tuple.RegionVersionConnectedRealmTuple,
						LastModified:                     time.Time{},
					},
					AuctionsResponse: AuctionsResponse{},
				}

				continue
			}

			auctionsResponse, responseMeta, err := NewAuctionsFromHTTP(getAuctionsUri, tuple.LastModified)
			if err != nil {
				out <- GetAuctionsJob{
					Err: err,
					Tuple: LoadConnectedRealmTuple{
						RegionVersionConnectedRealmTuple: tuple.RegionVersionConnectedRealmTuple,
						LastModified:                     time.Time{},
					},
					AuctionsResponse: AuctionsResponse{},
				}

				continue
			}

			if responseMeta.Status == http.StatusNotModified {
				out <- GetAuctionsJob{
					Err: nil,
					Tuple: LoadConnectedRealmTuple{
						RegionVersionConnectedRealmTuple: tuple.RegionVersionConnectedRealmTuple,
						LastModified:                     time.Time{},
					},
					AuctionsResponse: AuctionsResponse{},
				}

				continue
			}

			out <- GetAuctionsJob{
				Err: nil,
				Tuple: LoadConnectedRealmTuple{
					RegionVersionConnectedRealmTuple: tuple.RegionVersionConnectedRealmTuple,
					LastModified:                     responseMeta.LastModified,
				},
				AuctionsResponse: auctionsResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// spinning it up
	go func() {
		for _, tuple := range opts.Tuples {
			in <- tuple
		}

		close(in)
	}()

	return out
}
