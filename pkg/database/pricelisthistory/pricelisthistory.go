package pricelisthistory

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

// bucketing
func baseBucketName(ID blizzardv2.ItemId) []byte {
	return []byte(fmt.Sprintf("item-prices/%d", ID))
}

func recipeBucketName(id blizzardv2.RecipeId) []byte {
	return []byte(fmt.Sprintf("recipe-prices/%d", id))
}

// keying
func baseKeyName() []byte {
	return []byte("item-prices")
}

func recipeKeyName() []byte {
	return []byte("recipe-prices")
}

// db
func databaseFilePath(
	dirPath string,
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	targetTimestamp sotah.UnixTimestamp,
) string {
	return fmt.Sprintf("%s/%d.db", databaseDirPath(dirPath, tuple), targetTimestamp)
}

func databaseDirPath(
	dirPath string,
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) string {
	return fmt.Sprintf(
		"%s/pricelist-history/%s/%s/%d",
		dirPath,
		tuple.RegionName,
		tuple.Version,
		tuple.ConnectedRealmId,
	)
}
