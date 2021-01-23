package stats

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (rBases RegionDatabases) PersistEncodedStats(
	currentTimestamp sotah.UnixTimestamp,
	regionData map[blizzardv2.RegionName][]byte,
) error {
	for name, encodedData := range regionData {
		rBase, err := rBases.GetRegionDatabase(name)
		if err != nil {
			return err
		}

		if err := rBase.PersistEncodedStats(currentTimestamp, encodedData); err != nil {
			return err
		}
	}

	return nil
}
