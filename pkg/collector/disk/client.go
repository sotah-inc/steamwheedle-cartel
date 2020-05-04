package disk

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type ClientOptions struct {
	LakeClient      BaseLake.Client
	MessengerClient messenger.Messenger

	ResolveAuctions         func() chan blizzardv2.GetAuctionsJob
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}

func NewClient(opts ClientOptions) Client {
	return Client{
		lakeClient:              opts.LakeClient,
		messengerClient:         opts.MessengerClient,
		resolveAuctions:         opts.ResolveAuctions,
		receiveRegionTimestamps: opts.ReceiveRegionTimestamps,
	}
}

type Client struct {
	lakeClient      BaseLake.Client
	messengerClient messenger.Messenger

	resolveAuctions         func() chan blizzardv2.GetAuctionsJob
	receiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}
