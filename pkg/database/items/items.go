package items

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func baseBucketName() []byte {
	return []byte("items")
}

func namesBucketName() []byte {
	return []byte("item-names")
}

func blacklistBucketName() []byte {
	return []byte("item-blacklist")
}

func itemClassesBucket() []byte {
	return []byte("item-classes")
}

func itemClassItemsBucket() []byte {
	return []byte("item-class-items")
}

func itemVendorPricesBucket() []byte {
	return []byte("item-vendor-prices")
}

// keying
func baseKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-%d", id))
}

func nameKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-name-%d", id))
}

func blacklistKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-blacklist-%d", id))
}

func itemIdFromKeyName(key []byte) (blizzardv2.ItemId, error) {
	unparsedItemId, err := strconv.Atoi(string(key)[len("item-"):])
	if err != nil {
		return blizzardv2.ItemId(0), err
	}

	return blizzardv2.ItemId(unparsedItemId), nil
}

func itemIdFromNameKeyName(key []byte) (blizzardv2.ItemId, error) {
	unparsedItemId, err := strconv.Atoi(string(key)[len("item-name-"):])
	if err != nil {
		return blizzardv2.ItemId(0), err
	}

	return blizzardv2.ItemId(unparsedItemId), nil
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
