package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	DiskLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/disk"
)

type DiskAuctionsState struct {
	BlizzardState BlizzardState
	RegionsState  *RegionsState

	DiskLakeClient DiskLake.Client
	ItemsDatabase  database.ItemsDatabase
}
