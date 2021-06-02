package regions

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
)

func (rBase Database) GetTuples(
	version gameversion.GameVersion,
) ([]blizzardv2.RegionConnectedRealmTuple, error) {
	downloadTuples, err := rBase.GetDownloadTuples(version)
	if err != nil {
		return []blizzardv2.RegionConnectedRealmTuple{}, err
	}

	out := make([]blizzardv2.RegionConnectedRealmTuple, len(downloadTuples))
	for i, tuple := range downloadTuples {
		out[i] = tuple.RegionConnectedRealmTuple
	}

	return out, nil
}
