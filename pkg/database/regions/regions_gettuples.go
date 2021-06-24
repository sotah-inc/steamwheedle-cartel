package regions

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
)

func (rBase Database) GetTuples(
	version gameversion.GameVersion,
) (blizzardv2.RegionVersionConnectedRealmTuples, error) {
	downloadTuples, err := rBase.GetDownloadTuples()
	if err != nil {
		return blizzardv2.RegionVersionConnectedRealmTuples{}, err
	}

	out := make(blizzardv2.RegionVersionConnectedRealmTuples, len(downloadTuples))
	for i, tuple := range downloadTuples {
		if tuple.Version != version {
			continue
		}

		out[i] = blizzardv2.RegionVersionConnectedRealmTuple{
			RegionVersionTuple: blizzardv2.RegionVersionTuple{
				RegionTuple: blizzardv2.RegionTuple{RegionName: tuple.RegionName},
				Version:     tuple.Version,
			},
			ConnectedRealmId: tuple.ConnectedRealmId,
		}
	}

	return out, nil
}
