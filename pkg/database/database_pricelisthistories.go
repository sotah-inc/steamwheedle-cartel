package database

import (
	"encoding/binary"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/sotah"
)

// keying
func pricelistHistoryKeyName() []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, 1)

	return key
}

// bucketing
func pricelistHistoryBucketName(ID blizzard.ItemID) []byte {
	return []byte(fmt.Sprintf("item-prices/%d", ID))
}

// db
func pricelistHistoryDatabaseFilePath(
	dirPath string,
	regionName blizzard.RegionName,
	realmSlug blizzard.RealmSlug,
	targetTimestamp sotah.UnixTimestamp,
) string {
	return fmt.Sprintf(
		"%s/pricelist-histories/%s/%s/%d.db",
		dirPath,
		regionName,
		realmSlug,
		targetTimestamp,
	)
}
