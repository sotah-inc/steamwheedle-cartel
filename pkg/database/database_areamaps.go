package database

import (
	"fmt"
	"strconv"
)

// bucketing
func databaseAreaMapsBucketName() []byte {
	return []byte("area-maps")
}

func databaseAreaMapNamesBucketName() []byte {
	return []byte("area-map-names")
}

// keying
func areaMapsKeyName(id int) []byte {
	return []byte(fmt.Sprintf("area-map-%d", id))
}

func areaMapNameKeyName(id int) []byte {
	return []byte(fmt.Sprintf("area-map-name-%d", id))
}

func areaMapIdFromAreaMapNameKeyName(key []byte) (int, error) {
	unparsedItemId, err := strconv.Atoi(string(key)[len("area-map-name-"):])
	if err != nil {
		return 0, err
	}

	return unparsedItemId, nil
}

// db
func AreaMapsDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/areamaps.db", dbDir), nil
}
