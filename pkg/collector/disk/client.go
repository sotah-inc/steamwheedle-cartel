package disk

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type ClientOptions struct {
	LakeClient BaseLake.Client

	ResolveAuctions         func() chan blizzardv2.GetAuctionsJob
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}

func NewClient(opts ClientOptions) Client {
	return Client{
		lakeClient:              opts.LakeClient,
		resolveAuctions:         opts.ResolveAuctions,
		receiveRegionTimestamps: opts.ReceiveRegionTimestamps,
	}
}

type Client struct {
	lakeClient BaseLake.Client

	resolveAuctions         func() chan blizzardv2.GetAuctionsJob
	receiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}
