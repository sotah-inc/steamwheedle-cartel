package database

import (
	"fmt"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func liveAuctionsDatabasePath(dirPath string, tuple blizzardv2.RegionConnectedRealmTuple) string {
	return fmt.Sprintf("%s/live-auctions/%s/%d.db", dirPath, tuple.RegionName, tuple.ConnectedRealmId)
}

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

func normalizeLiveAuctionsStatsLastUpdated(lastUpdatedTimestamp int64) int64 {
	lastUpdated := time.Unix(lastUpdatedTimestamp, 0)
	nearestHourOffset := lastUpdated.Second() + lastUpdated.Minute()*60

	return time.Unix(lastUpdatedTimestamp-int64(nearestHourOffset), 0).Unix()
}

func unixTimestampFromLiveAuctionsStatsKeyName(key []byte) (int64, error) {
	decodedLastUpdated, err := strconv.Atoi(string(key)[len("live-auctions-stats-"):])
	if err != nil {
		return int64(0), err
	}

	return int64(decodedLastUpdated), nil
}
