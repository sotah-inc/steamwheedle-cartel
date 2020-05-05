package disk

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

type NewClientOptions struct {
	CacheDir          string
	ResolveItems      func(ids blizzardv2.ItemIds) chan blizzardv2.GetItemsOutJob
	ResolveItemMedias func(in chan blizzardv2.GetItemMediasInJob) chan blizzardv2.GetItemMediasOutJob
}

func NewClient(opts NewClientOptions) Client {
	return Client{
		cacheDir:          opts.CacheDir,
		resolveItems:      opts.ResolveItems,
		resolveItemMedias: opts.ResolveItemMedias,
	}
}

type Client struct {
	cacheDir          string
	resolveItems      func(ids blizzardv2.ItemIds) chan blizzardv2.GetItemsOutJob
	resolveItemMedias func(in chan blizzardv2.GetItemMediasInJob) chan blizzardv2.GetItemMediasOutJob
}
