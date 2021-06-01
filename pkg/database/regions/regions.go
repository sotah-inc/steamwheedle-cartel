package regions

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func baseBucketName() []byte {
	return []byte("regions")
}

func connectedRealmsBucketName(version gameversion.GameVersion, name blizzardv2.RegionName) []byte {
	return []byte(fmt.Sprintf("connected-realms-%s-%s", version, name))
}

// base keying
func baseKeyName(name blizzardv2.RegionName) []byte {
	return []byte(fmt.Sprintf("region-%s", name))
}

func regionNameFromKeyName(key []byte) blizzardv2.RegionName {
	return blizzardv2.RegionName(string(key)[len("region-"):])
}

// realms keying
func connectedRealmsKeyName(id blizzardv2.ConnectedRealmId) []byte {
	return []byte(fmt.Sprintf("connected-realm-%d", id))
}

func connectedRealmIdFromKeyName(key []byte) (blizzardv2.ConnectedRealmId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("connected-realm-"):])
	if err != nil {
		return blizzardv2.ConnectedRealmId(0), err
	}

	return blizzardv2.ConnectedRealmId(unparsedId), nil
}

// db
func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/regions.db", dbDir), nil
}
