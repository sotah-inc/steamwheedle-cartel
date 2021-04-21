package items

import (
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

// db
func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/items.db", dbDir), nil
}
