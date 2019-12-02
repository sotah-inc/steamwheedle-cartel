package database

import (
	"fmt"

	"git.sotah.info/steamwheedle-cartel/pkg/sotah"
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
