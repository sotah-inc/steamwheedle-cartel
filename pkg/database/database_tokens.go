package database

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
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
	decodedLastUpdated, err := strconv.Atoi(string(key)[len("last-updated-"):])
	if err != nil {
		return int64(0), err
	}

	return int64(decodedLastUpdated), nil
}

func priceFromTokenValue(v []byte) (int64, error) {
	decodedValue, err := strconv.Atoi(string(v))
	if err != nil {
		return int64(0), err
	}

	return int64(decodedValue), nil
}

// db
func TokensDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/tokens.db", dbDir), nil
}
