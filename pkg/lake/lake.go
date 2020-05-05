package lake

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	DiskLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/disk"
)

type NewClientOptions struct {
	UseGCloud         bool
	CacheDir          string
	ResolveItems      func(ids blizzardv2.ItemIds) chan blizzardv2.GetItemsOutJob
	ResolveItemMedias func(in chan blizzardv2.GetItemMediasInJob) chan blizzardv2.GetItemMediasOutJob
}

func NewClient(opts NewClientOptions) (BaseLake.Client, error) {
	if opts.UseGCloud {
		return DiskLake.NewClient(DiskLake.NewClientOptions{
			CacheDir:          opts.CacheDir,
			ResolveItems:      opts.ResolveItems,
			ResolveItemMedias: opts.ResolveItemMedias,
		})
	}

	return DiskLake.NewClient(DiskLake.NewClientOptions{
		CacheDir:          opts.CacheDir,
		ResolveItems:      opts.ResolveItems,
		ResolveItemMedias: opts.ResolveItemMedias,
	})
}
