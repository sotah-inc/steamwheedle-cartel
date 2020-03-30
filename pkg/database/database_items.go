package database

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func databaseItemsBucketName() []byte {
	return []byte("items")
}

func databaseItemNamesBucketName() []byte {
	return []byte("item-names")
}

// keying
func itemsKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-%d", id))
}

func itemNameKeyName(id blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-name-%d", id))
}

func itemIdFromItemNameKeyName(key []byte) (blizzardv2.ItemId, error) {
	unparsedItemId, err := strconv.Atoi(string(key)[len("item-name-"):])
	if err != nil {
		return blizzardv2.ItemId(0), err
	}

	return blizzardv2.ItemId(unparsedItemId), nil
}

// db
func ItemsDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/items.db", dbDir), nil
}
