package pricelisthistory

import (
	"encoding/binary"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

// keying
func pricelistHistoryKeyName() []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, 1)

	return key
}

// bucketing
func pricelistHistoryBucketName(ID blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-prices/%d", ID))
}

// db
func pricelistHistoryDatabaseFilePath(
	dirPath string,
	tuple blizzardv2.RegionConnectedRealmTuple,
	targetTimestamp sotah.UnixTimestamp,
) string {
	return fmt.Sprintf(
		"%s/pricelist-history/%s/%d/%d.db",
		dirPath,
		tuple.RegionName,
		tuple.ConnectedRealmId,
		targetTimestamp,
	)
}
