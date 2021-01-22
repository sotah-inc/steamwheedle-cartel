package stats

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func newRegionDatabase(dirPath string, name blizzardv2.RegionName) (RegionDatabase, error) {
	db, err := newDatabase(regionDatabasePath(dirPath, name))
	if err != nil {
		return RegionDatabase{}, err
	}

	return RegionDatabase{db, name}, nil
}

type RegionDatabase struct {
	Database
	name blizzardv2.RegionName
}
