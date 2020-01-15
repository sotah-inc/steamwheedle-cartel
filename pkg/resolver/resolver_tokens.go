package resolver

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (r Resolver) NewTokenInfo(regionHostname string) (blizzard.TokenInfo, error) {
	resp, err := r.Download(r.GetTokenInfoURL(regionHostname), true)
	if err != nil {
		return blizzard.TokenInfo{}, err
	}
	if resp.Status != http.StatusOK {
		return blizzard.TokenInfo{}, errors.New("response when downloading token info was not OK")
	}

	return blizzard.NewTokenInfo(resp.Body)
}

type GetTokensJob struct {
	Err   error
	Tuple RegionHostnameTuple
	Info  blizzard.TokenInfo
}

func (job GetTokensJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":    job.Err.Error(),
		"region":   job.Tuple.RegionName,
		"hostname": job.Tuple.Hostname,
	}
}

type RegionHostnameTuple struct {
	RegionName blizzard.RegionName
	Hostname   string
}

func (r Resolver) GetTokens(tuples []RegionHostnameTuple) chan GetTokensJob {
	// establishing channels
	out := make(chan GetTokensJob)
	in := make(chan RegionHostnameTuple)

	// spinning up the workers for fetching items
	worker := func() {
		for tuple := range in {
			tInfo, err := r.NewTokenInfo(tuple.Hostname)
			out <- GetTokensJob{err, tuple, tInfo}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing up the items
	go func() {
		for _, tuple := range tuples {
			in <- tuple
		}

		close(in)
	}()

	return out
}
