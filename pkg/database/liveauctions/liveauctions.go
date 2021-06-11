package liveauctions

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func databasePath(dirPath string, tuple blizzardv2.RegionVersionConnectedRealmTuple) string {
	return fmt.Sprintf(
		"%s/live-auctions/%s/%s/%d.db",
		dirPath,
		tuple.RegionName,
		tuple.Version,
		tuple.ConnectedRealmId,
	)
}

// bucket, key for current live-auctions
func baseBucketName() []byte {
	return []byte("live-auctions")
}

func baseKeyName() []byte {
	return []byte("live-auctions")
}
