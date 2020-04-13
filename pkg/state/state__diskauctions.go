package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
)

type DiskAuctionsState struct {
	BlizzardState BlizzardState
	RegionsState  *RegionsState

	DiskStore     diskstore.DiskStore
	ItemsDatabase database.ItemsDatabase
}
