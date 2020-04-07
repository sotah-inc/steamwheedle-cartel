package diskstore

import (
	"errors"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (ds DiskStore) resolveAuctionsFilepath(tuple blizzardv2.RegionConnectedRealmTuple) (string, error) {
	if len(ds.CacheDir) == 0 {
		return "", errors.New("cache dir cannot be blank")
	}

	return fmt.Sprintf("%s/auctions/%s/%d.json.gz", ds.CacheDir, tuple.RegionName, tuple.ConnectedRealmId), nil
}
