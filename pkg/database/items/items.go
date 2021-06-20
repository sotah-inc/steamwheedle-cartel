package items

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
)

// bucketing

func baseBucketName(version gameversion.GameVersion) []byte {
	return []byte(fmt.Sprintf("items-%s", version))
}

func namesBucketName(version gameversion.GameVersion) []byte {
	return []byte(fmt.Sprintf("item-names-%s", version))
}

func blacklistBucketName(version gameversion.GameVersion) []byte {
	return []byte(fmt.Sprintf("item-blacklist-%s", version))
}

func itemClassesBucket() []byte {
	return []byte("item-classes")
}

func itemClassItemsBucket() []byte {
	return []byte("item-class-items")
}

func itemVendorPricesBucket(version gameversion.GameVersion) []byte {
	return []byte(fmt.Sprintf("item-vendor-prices-%s", version))
}

// keying

func baseKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-%d", id))
}

func itemIdFromKeyName(key []byte) (blizzardv2.ItemId, error) {
	unparsedItemId, err := strconv.Atoi(string(key)[len("item-"):])
	if err != nil {
		return blizzardv2.ItemId(0), err
	}

	return blizzardv2.ItemId(unparsedItemId), nil
}

func nameKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-name-%d", id))
}

func itemIdFromNameKeyName(key []byte) (blizzardv2.ItemId, error) {
	unparsedItemId, err := strconv.Atoi(string(key)[len("item-name-"):])
	if err != nil {
		return blizzardv2.ItemId(0), err
	}

	return blizzardv2.ItemId(unparsedItemId), nil
}

func blacklistKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-blacklist-%d", id))
}

func itemIdFromBlacklistKeyName(key []byte) (blizzardv2.ItemId, error) {
	unparsedItemId, err := strconv.Atoi(string(key)[len("item-blacklist-"):])
	if err != nil {
		return blizzardv2.ItemId(0), err
	}

	return blizzardv2.ItemId(unparsedItemId), nil
}

func itemClassesKeyName() []byte {
	return []byte("item-classes")
}

func itemClassItemsKeyName(id itemclass.Id) []byte {
	return []byte(fmt.Sprintf("item-class-%d-item-ids", id))
}

func itemVendorPriceKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-%d-vendor-price", id))
}

func itemVendorPriceToValue(v blizzardv2.PriceValue) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))

	return b
}

func itemVendorPriceFromValue(v []byte) blizzardv2.PriceValue {
	return blizzardv2.PriceValue(int64(binary.LittleEndian.Uint64(v)))
}

// db

func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/items.db", dbDir), nil
}
