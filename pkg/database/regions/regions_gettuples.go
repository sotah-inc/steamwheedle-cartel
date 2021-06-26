package regions

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (rBase Database) GetTuples() (blizzardv2.RegionVersionConnectedRealmTuples, error) {
	downloadTuples, err := rBase.GetDownloadTuples()
	if err != nil {
		return blizzardv2.RegionVersionConnectedRealmTuples{}, err
	}

	out := make(blizzardv2.RegionVersionConnectedRealmTuples, len(downloadTuples))
	for i, tuple := range downloadTuples {
		out[i] = tuple.RegionVersionConnectedRealmTuple
	}

	return out, nil
}
