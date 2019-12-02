package database

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/sotah"
)

func liveAuctionsBucketName() []byte {
	return []byte("live-auctions")
}

func liveAuctionsKeyName() []byte {
	return []byte("live-auctions")
}

func liveAuctionsDatabasePath(dirPath string, rea sotah.Realm) string {
	return fmt.Sprintf("%s/live-auctions/%s/%s.db", dirPath, rea.Region.Name, rea.Slug)
}
