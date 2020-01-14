package database

import (
	"fmt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"strconv"
)

// bucketing
func databaseTokensBucketName(regionName blizzard.RegionName) []byte {
	return []byte(fmt.Sprintf("tokens-%s", regionName))
}

// keying
func tokenKeyName(lastUpdated int64) []byte {
	return []byte(fmt.Sprintf("last-updated-%d", lastUpdated))
}


func lastUpdatedFromTokenKeyName(key []byte) (int64, error) {
	unparsedLastUpdated, err := strconv.Atoi(string(key)[len("last-updated-"):])
	if err != nil {
		return int64(0), err
	}

	return int64(unparsedLastUpdated), nil
}

// db
func TokenDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/tokens.db", dbDir), nil
}
