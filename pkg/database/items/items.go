package items

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
)

// bucketing

func baseBucketName() []byte {
	return []byte("items")
}

func namesBucketName() []byte {
	return []byte("item-names")
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

func baseKeyName(tuple blizzardv2.VersionItemTuple) []byte {
	return []byte(fmt.Sprintf("item/%s-%d", tuple.GameVersion, tuple.Id))
}

func tupleFromBaseKeyName(key []byte) (blizzardv2.VersionItemTuple, error) {
	slashParts := strings.Split(string(key), "/")
	if len(slashParts) != 2 {
		return blizzardv2.VersionItemTuple{}, errors.New("base key name had incorrect length")
	}

	dashParts := strings.Split(slashParts[1], "-")
	if len(dashParts) != 2 {
		return blizzardv2.VersionItemTuple{}, errors.New("base key name had incorrect length")
	}

	parsedItemId, err := strconv.Atoi(dashParts[2])
	if err != nil {
		return blizzardv2.VersionItemTuple{}, err
	}

	return blizzardv2.VersionItemTuple{
		GameVersion: gameversion.GameVersion(dashParts[1]),
		Id:          blizzardv2.ItemId(parsedItemId),
	}, nil
}

func nameKeyName(tuple blizzardv2.VersionItemTuple) []byte {
	return []byte(fmt.Sprintf("item-name/%s-%d", tuple.GameVersion, tuple.Id))
}

func tupleFromNameKeyName(key []byte) (blizzardv2.VersionItemTuple, error) {
	return tupleFromBaseKeyName(key)
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
