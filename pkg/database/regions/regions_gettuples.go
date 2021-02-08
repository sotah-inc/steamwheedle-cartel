package regions

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) GetTuples() ([]blizzardv2.RegionConnectedRealmTuple, error) {
	downloadTuples, err := rBase.GetDownloadTuples()
	if err != nil {
		return []blizzardv2.RegionConnectedRealmTuple{}, err
	}

	out := make([]blizzardv2.RegionConnectedRealmTuple, len(downloadTuples))
	for i, tuple := range downloadTuples {
		out[i] = tuple.RegionConnectedRealmTuple
	}

	return out, nil
}
