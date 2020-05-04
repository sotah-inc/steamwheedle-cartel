package lake

import (
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	DiskLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/disk"
)

func NewClient(useGCloud bool, cacheDir string) BaseLake.Client {
	if useGCloud {
		return DiskLake.NewClient(cacheDir)
	}

	return DiskLake.NewClient(cacheDir)
}
