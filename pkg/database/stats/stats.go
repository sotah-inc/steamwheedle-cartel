package stats

import (
	"fmt"
	"strconv"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func tupleDatabasePath(dirPath string, tuple blizzardv2.RegionVersionConnectedRealmTuple) string {
	return databasePath(
		dirPath,
		fmt.Sprintf("%s/%s/%d", tuple.RegionName, tuple.Version, tuple.ConnectedRealmId),
	)
}

func regionDatabasePath(dirPath string, name blizzardv2.RegionName) string {
	return databasePath(dirPath, string(name))
}

func databasePath(dirPath string, suffix string) string {
	return fmt.Sprintf("%s/stats/%s.db", dirPath, suffix)
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

	return sotah.UnixTimestamp(
		time.Unix(int64(lastUpdatedTimestamp)-int64(nearestHourOffset), 0).Unix(),
	)
}

func unixTimestampFromBaseKeyName(key []byte) (sotah.UnixTimestamp, error) {
	decodedLastUpdated, err := strconv.Atoi(string(key)[len("stats-"):])
	if err != nil {
		return 0, err
	}

	return sotah.UnixTimestamp(decodedLastUpdated), nil
}
