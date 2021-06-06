package disk

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type ClientOptions struct {
	LakeClient      BaseLake.Client
	MessengerClient messenger.Messenger

	ResolveAuctions func(
		version gameversion.GameVersion,
	) (chan blizzardv2.GetAuctionsJob, error)
	ReceiveRegionTimestamps func(
		version gameversion.GameVersion,
		timestamps sotah.RegionTimestamps,
	) error
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

	resolveAuctions func(
		version gameversion.GameVersion,
	) (chan blizzardv2.GetAuctionsJob, error)
	receiveRegionTimestamps func(
		version gameversion.GameVersion,
		timestamps sotah.RegionTimestamps,
	) error
}
