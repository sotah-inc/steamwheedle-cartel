package disk

import (
	"errors"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (client Client) resolveAuctionsFilepath(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) (string, error) {
	if len(client.cacheDir) == 0 {
		return "", errors.New("cache dir cannot be blank")
	}

	return fmt.Sprintf(
		"%s/auctions/%s/%s/%d.json.gz",
		client.cacheDir,
		tuple.RegionName,
		tuple.Version,
		tuple.ConnectedRealmId,
	), nil
}
