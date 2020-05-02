package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type DiskAuctionsState struct {
	BlizzardState           BlizzardState
	GetTuples               func() []blizzardv2.DownloadConnectedRealmTuple
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)

	LakeClient BaseLake.Client
}
