package database

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucket, key for current live-auctions
func liveAuctionsBucketName() []byte {
	return []byte("live-auctions")
}

func liveAuctionsMainKeyName() []byte {
	return []byte("live-auctions")
}

func liveAuctionsDatabasePath(dirPath string, tuple blizzardv2.RegionConnectedRealmTuple) string {
	return fmt.Sprintf("%s/live-auctions/%s/%d.db", dirPath, tuple.RegionName, tuple.ConnectedRealmId)
}
