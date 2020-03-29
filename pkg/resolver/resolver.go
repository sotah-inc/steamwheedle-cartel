package resolver

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
)

func NewResolver(bc blizzardv2.Client, re metric.Reporter) Resolver {
	return Resolver{
		BlizzardClient: bc,
		Reporter:       re,

		GetItemURL:           blizzardv2.DefaultGetItemURL,
		GetItemIconURL:       blizzardv2.DefaultGetItemIconURL,
		GetItemClassIndexURL: blizzardv2.DefaultGetItemClassIndexURL,
		GetItemClassURL:      blizzardv2.DefaultGetItemClassURL,
		GetTokenInfoURL:      blizzardv2.DefaultGetTokenInfoURL,
	}
}

type Resolver struct {
	BlizzardClient blizzardv2.Client
	Reporter       metric.Reporter

	GetItemURL           blizzardv2.GetItemURLFunc
	GetItemIconURL       blizzardv2.GetItemIconURLFunc
	GetItemClassIndexURL blizzardv2.GetItemClassIndexURLFunc
	GetItemClassURL      blizzardv2.GetItemClassURLFunc
	GetTokenInfoURL      blizzardv2.GetTokenInfoURLFunc
}

func (r Resolver) AppendAccessToken(destination string) (string, error) {
	return r.BlizzardClient.AppendAccessToken(destination)
}

func (r Resolver) Download(uri string, shouldAppendAccessToken bool) (blizzardv2.ResponseMeta, error) {
	uri, err := func() (string, error) {
		if !shouldAppendAccessToken {
			return uri, nil
		}

		return r.AppendAccessToken(uri)
	}()
	if err != nil {
		return blizzardv2.ResponseMeta{}, err
	}

	resp, err := blizzardv2.Download(blizzardv2.DownloadOptions{Uri: uri})
	if resp.RequestDuration > 0 || resp.ConnectionDuration > 0 {
		r.Reporter.Report(metric.Metrics{
			"conn_duration":    int(resp.ConnectionDuration / 1000 / 1000),
			"request_duration": int(resp.RequestDuration / 1000 / 1000),
		})
	}

	if err != nil {
		return blizzardv2.ResponseMeta{}, err
	}

	return resp, nil
}
