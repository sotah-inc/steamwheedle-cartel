package database

import (
	"fmt"
	"strconv"
)

// bucketing
func databaseTokensBucketName() []byte {
	return []byte("tokens")
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
