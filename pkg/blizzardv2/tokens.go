package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetTokensJob struct {
	Err   error
	Tuple RegionHostnameTuple
	Token TokenResponse
}

func (job GetTokensJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":    job.Err.Error(),
		"region":   job.Tuple.RegionName,
		"hostname": job.Tuple.RegionHostname,
	}
}

type RegionHostnameTuple struct {
	RegionName     RegionName
	RegionHostname string
}

type GetTokensOptions struct {
	Tuples          []RegionHostnameTuple
	GetTokenInfoURL func(string, RegionName) (string, error)
}

func GetTokens(opts GetTokensOptions) (map[RegionName]TokenResponse, error) {
	// starting up workers for gathering individual tokens
	in := make(chan RegionHostnameTuple)
	out := make(chan GetTokensJob)
	worker := func() {
		for tuple := range in {
			getTokenUri, err := opts.GetTokenInfoURL(tuple.RegionHostname, tuple.RegionName)
			if err != nil {
				out <- GetTokensJob{
					Err:   err,
					Tuple: tuple,
				}

				continue
			}

			tokenResponse, _, err := NewTokenFromHTTP(getTokenUri)
			if err != nil {
				out <- GetTokensJob{
					Err:   err,
					Tuple: tuple,
				}

				continue
			}

			out <- GetTokensJob{
				Err:   nil,
				Tuple: tuple,
				Token: tokenResponse,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		for _, tuple := range opts.Tuples {
			in <- tuple
		}

		close(in)
	}()

	// waiting for it to drain out
	results := map[RegionName]TokenResponse{}
	for job := range out {
		if job.Err != nil {
			return map[RegionName]TokenResponse{}, job.Err
		}

		results[job.Tuple.RegionName] = job.Token
	}

	return results, nil
}
