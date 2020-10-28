package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetAllConnectedRealmsOptions struct {
	GetConnectedRealmIndexURL func() (string, error)
	GetConnectedRealmURL      func(string) (string, error)
}

type GetAllConnectedRealmsJob struct {
	Err                    error
	HrefReference          HrefReference
	ConnectedRealmResponse ConnectedRealmResponse
}

func (job GetAllConnectedRealmsJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"href":  job.HrefReference.Href,
	}
}

func GetAllConnectedRealms(opts GetAllConnectedRealmsOptions) ([]ConnectedRealmResponse, error) {
	// querying index
	uri, err := opts.GetConnectedRealmIndexURL()
	if err != nil {
		return []ConnectedRealmResponse{}, err
	}

	crIndex, _, err := NewConnectedRealmIndexFromHTTP(uri)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get connected-realm-index")

		return []ConnectedRealmResponse{}, err
	}

	// starting up workers for gathering individual connected-realms
	in := make(chan HrefReference)
	out := make(chan GetAllConnectedRealmsJob)
	worker := func() {
		for hrefRef := range in {
			getConnectedRealmUri, err := opts.GetConnectedRealmURL(hrefRef.Href)
			if err != nil {
				out <- GetAllConnectedRealmsJob{
					Err:                    err,
					HrefReference:          hrefRef,
					ConnectedRealmResponse: ConnectedRealmResponse{},
				}

				continue
			}

			cRealm, _, err := NewConnectedRealmResponseFromHTTP(getConnectedRealmUri)
			if err != nil {
				out <- GetAllConnectedRealmsJob{
					Err:                    err,
					HrefReference:          hrefRef,
					ConnectedRealmResponse: ConnectedRealmResponse{},
				}

				continue
			}

			out <- GetAllConnectedRealmsJob{
				Err:                    nil,
				HrefReference:          hrefRef,
				ConnectedRealmResponse: cRealm,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing it up
	go func() {
		for _, hrefRef := range crIndex.ConnectedRealms {
			in <- hrefRef
		}

		close(in)
	}()

	// waiting for it all to drain out
	result := make([]ConnectedRealmResponse, len(crIndex.ConnectedRealms))
	i := 0
	for outJob := range out {
		if outJob.Err != nil {
			return []ConnectedRealmResponse{}, outJob.Err
		}

		result[i] = outJob.ConnectedRealmResponse
		i += 1
	}

	return result, nil
}
