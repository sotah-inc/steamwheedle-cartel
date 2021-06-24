package regions

import (
	"fmt"
	"strconv"
	"strings"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func baseBucketName() []byte {
	return []byte("regions")
}

func connectedRealmsBucketName() []byte {
	return []byte("connected-realms")
}

// base keying
func baseKeyName(name blizzardv2.RegionName) []byte {
	return []byte(fmt.Sprintf("region-%s", name))
}

func regionNameFromKeyName(key []byte) blizzardv2.RegionName {
	return blizzardv2.RegionName(string(key)[len("region-"):])
}

// realms keying
func connectedRealmsKeyName(tuple blizzardv2.RegionVersionConnectedRealmTuple) []byte {
	return []byte(
		fmt.Sprintf("%s/%s/%d", tuple.RegionName, tuple.Version, tuple.ConnectedRealmId),
	)
}

func tupleFromConnectedRealmKeyName(
	key []byte,
) (blizzardv2.RegionVersionConnectedRealmTuple, error) {
	parts := strings.Split(string(key), "/")

	connectedRealmId, err := strconv.Atoi(parts[2])
	if err != nil {
		return blizzardv2.RegionVersionConnectedRealmTuple{}, err
	}

	return blizzardv2.RegionVersionConnectedRealmTuple{
		RegionVersionTuple: blizzardv2.RegionVersionTuple{
			RegionTuple: blizzardv2.RegionTuple{
				RegionName: blizzardv2.RegionName(parts[0]),
			},
			Version: gameversion.GameVersion(parts[1]),
		},
		ConnectedRealmId: blizzardv2.ConnectedRealmId(connectedRealmId),
	}, nil
}

// db

func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/regions.db", dbDir), nil
}
