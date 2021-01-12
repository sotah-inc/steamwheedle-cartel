package stats

import (
	"fmt"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func databasePath(dirPath string, tuple blizzardv2.RegionConnectedRealmTuple) string {
	return fmt.Sprintf("%s/stats/%s/%d.db", dirPath, tuple.RegionName, tuple.ConnectedRealmId)
}

// bucket, key for current live-auctions
func baseBucketName() []byte {
	return []byte("stats")
}

func baseKeyName(lastUpdated sotah.UnixTimestamp) []byte {
	return []byte(fmt.Sprintf("stats-%d", lastUpdated))
}

func normalizeLastUpdated(lastUpdatedTimestamp sotah.UnixTimestamp) sotah.UnixTimestamp {
	lastUpdated := time.Unix(int64(lastUpdatedTimestamp), 0)
	nearestHourOffset := lastUpdated.Second() + lastUpdated.Minute()*60

	return sotah.UnixTimestamp(time.Unix(int64(lastUpdatedTimestamp)-int64(nearestHourOffset), 0).Unix())
}

func unixTimestampFromBaseKeyName(key []byte) (sotah.UnixTimestamp, error) {
	decodedLastUpdated, err := strconv.Atoi(string(key)[len("stats-"):])
	if err != nil {
		return 0, err
	}

	return sotah.UnixTimestamp(decodedLastUpdated), nil
}
