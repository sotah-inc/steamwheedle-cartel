package tokens

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func baseBucketName(regionName blizzardv2.RegionName) []byte {
	return []byte(fmt.Sprintf("tokens-%s", regionName))
}

// keying
func baseKeyName(lastUpdated int64) []byte {
	return []byte(fmt.Sprintf("last-updated-%d", lastUpdated))
}

func lastUpdatedFromBaseKeyName(key []byte) (int64, error) {
	decodedLastUpdated, err := strconv.Atoi(string(key)[len("last-updated-"):])
	if err != nil {
		return int64(0), err
	}

	return int64(decodedLastUpdated), nil
}

func priceToTokenValue(v int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))

	return b
}

func priceFromTokenValue(v []byte) int64 {
	return int64(binary.LittleEndian.Uint64(v))
}

// db
func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/tokens.db", dbDir), nil
}
