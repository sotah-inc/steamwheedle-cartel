package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
)

type LiveAuctionsState struct {
	LiveAuctionsDatabases database.LiveAuctionsDatabases

	Messenger messenger.Messenger
}
