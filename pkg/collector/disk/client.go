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

	ResolveAuctions         func(tuples []blizzardv2.DownloadConnectedRealmTuple) chan blizzardv2.GetAuctionsJob
	GetTuples               func() []blizzardv2.DownloadConnectedRealmTuple
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}

func NewClient(opts ClientOptions) Client {
	return Client{
		lakeClient:              opts.LakeClient,
		messengerClient:         opts.MessengerClient,
		resolveAuctions:         opts.ResolveAuctions,
		getTuples:               opts.GetTuples,
		receiveRegionTimestamps: opts.ReceiveRegionTimestamps,
	}
}

type Client struct {
	lakeClient      BaseLake.Client
	messengerClient messenger.Messenger

	resolveAuctions         func(tuples []blizzardv2.DownloadConnectedRealmTuple) chan blizzardv2.GetAuctionsJob
	getTuples               func() []blizzardv2.DownloadConnectedRealmTuple
	receiveRegionTimestamps func(timestamps sotah.RegionTimestamps)
}
