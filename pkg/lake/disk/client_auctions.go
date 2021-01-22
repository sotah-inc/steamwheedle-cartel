package disk

import (
	"errors"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (client Client) resolveAuctionsFilepath(
	tuple blizzardv2.RegionConnectedRealmTuple,
) (string, error) {
	if len(client.cacheDir) == 0 {
		return "", errors.New("cache dir cannot be blank")
	}

	return fmt.Sprintf(
		"%s/auctions/%s/%d.json.gz",
		client.cacheDir,
		tuple.RegionName,
		tuple.ConnectedRealmId,
	), nil
}
