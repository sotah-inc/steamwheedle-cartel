package database

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

// bucket, key for current live-auctions
func liveAuctionsBucketName() []byte {
	return []byte("live-auctions")
}

func liveAuctionsMainKeyName() []byte {
	return []byte("live-auctions")
}

// bucket, key for live-auctions stats
func liveAuctionsStatsBucketName() []byte {
	return []byte("live-auctions-stats")
}

func liveAuctionsStatsKeyName(lastUpdated int64) []byte {
	return []byte(fmt.Sprintf("live-auctions-stats-%d", lastUpdated))
}

func unixTimestampFromLiveAuctionsStatsKeyName(key []byte) (int64, error) {
	decodedLastUpdated, err := strconv.Atoi(string(key)[len("live-auctions-stats-"):])
	if err != nil {
		return int64(0), err
	}

	return int64(decodedLastUpdated), nil
}

func liveAuctionsDatabasePath(dirPath string, rea sotah.Realm) string {
	return fmt.Sprintf("%s/live-auctions/%s/%s.db", dirPath, rea.Region.Name, rea.Slug)
}
