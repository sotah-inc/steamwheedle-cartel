package items

import (
	"fmt"
	"strconv"

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

// db
func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/items.db", dbDir), nil
}
