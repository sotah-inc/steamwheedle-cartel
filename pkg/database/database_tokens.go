package database

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
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
		logging.WithFields(logrus.Fields{
			"error":      err.Error(),
			"key":        key,
			"key-string": string(key),
		}).Error("Failed to convert last-updated key to integer")

		return int64(0), err
	}

	return int64(decodedLastUpdated), nil
}

func priceFromTokenValue(v []byte) (int64, error) {
	decodedValue, err := strconv.Atoi(string(v))
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":    err.Error(),
			"v":        v,
			"v-string": string(v),
		}).Error("Failed to convert price value to integer")

		return int64(0), err
	}

	return int64(decodedValue), nil
}

// db
func TokensDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/tokens.db", dbDir), nil
}
