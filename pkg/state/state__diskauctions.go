package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	DiskLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/disk"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type DiskAuctionsState struct {
	BlizzardState           BlizzardState
	GetTuples               func() []blizzardv2.DownloadConnectedRealmTuple
	ReceiveRegionTimestamps func(timestamps sotah.RegionTimestamps)

	DiskLakeClient DiskLake.Client
	ItemsDatabase  database.ItemsDatabase
}
