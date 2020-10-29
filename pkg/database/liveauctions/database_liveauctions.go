package liveauctions

import (
	"fmt"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func databasePath(dirPath string, tuple blizzardv2.RegionConnectedRealmTuple) string {
	return fmt.Sprintf("%s/live-auctions/%s/%d.db", dirPath, tuple.RegionName, tuple.ConnectedRealmId)
}

// bucket, key for current live-auctions
func baseBucketName() []byte {
	return []byte("live-auctions")
}

func baseKeyName() []byte {
	return []byte("live-auctions")
}

// bucket, key for live-auctions stats
func statsBucketName() []byte {
	return []byte("live-auctions-stats")
}

func statsKeyName(lastUpdated sotah.UnixTimestamp) []byte {
	return []byte(fmt.Sprintf("live-auctions-stats-%d", lastUpdated))
}

func normalizeLiveAuctionsStatsLastUpdated(lastUpdatedTimestamp sotah.UnixTimestamp) sotah.UnixTimestamp {
	lastUpdated := time.Unix(int64(lastUpdatedTimestamp), 0)
	nearestHourOffset := lastUpdated.Second() + lastUpdated.Minute()*60

	return sotah.UnixTimestamp(time.Unix(int64(lastUpdatedTimestamp)-int64(nearestHourOffset), 0).Unix())
}

func unixTimestampFromLiveAuctionsStatsKeyName(key []byte) (sotah.UnixTimestamp, error) {
	decodedLastUpdated, err := strconv.Atoi(string(key)[len("live-auctions-stats-"):])
	if err != nil {
		return 0, err
	}

	return sotah.UnixTimestamp(decodedLastUpdated), nil
}
