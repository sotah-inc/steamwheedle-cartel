package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ConnectedRealmResponses []ConnectedRealmResponse

type GetAuctionsJob struct {
	Err              error
	HrefReference    HrefReference
	AuctionsResponse AuctionsResponse
}

func (job GetAuctionsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"href":  job.HrefReference.Href,
	}
}

func (response ConnectedRealmResponses) GetAuctions(getAuctionsURL func(string) (string, error)) chan GetAuctionsJob {
	// starting up workers for gathering individual connected-realms
	in := make(chan HrefReference)
	out := make(chan GetAuctionsJob)
	worker := func() {
		for hrefRef := range in {
			getAuctionsUri, err := getAuctionsURL(hrefRef.Href)
			if err != nil {
				out <- GetAuctionsJob{
					Err:           err,
					HrefReference: hrefRef,
				}

				continue
			}

			auctionsResponse, _, err := NewAuctionsFromHTTP(getAuctionsUri)
			if err != nil {
				out <- GetAuctionsJob{
					Err:           err,
					HrefReference: hrefRef,
				}

				continue
			}

			out <- GetAuctionsJob{
				Err:              nil,
				HrefReference:    hrefRef,
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
