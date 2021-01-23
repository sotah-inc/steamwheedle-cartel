package stats

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewRegionDatabases(dirPath string, names []blizzardv2.RegionName) (RegionDatabases, error) {
	rBases := RegionDatabases{}

	for _, name := range names {
		rBase, err := newRegionDatabase(dirPath, name)
		if err != nil {
			return RegionDatabases{}, err
		}

		rBases[name] = rBase
	}

	return rBases, nil
}

type RegionDatabases map[blizzardv2.RegionName]RegionDatabase

func (rBases RegionDatabases) GetRegionDatabase(
	name blizzardv2.RegionName,
) (RegionDatabase, error) {
	db, ok := rBases[name]
	if !ok {
		return RegionDatabase{}, fmt.Errorf("db not found for region %s", name)
	}

	return db, nil
}
