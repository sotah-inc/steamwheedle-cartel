package sotah

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

func NewItemSyncWhitelist(ids blizzardv2.ItemIds) ItemSyncWhitelist {
	out := ItemSyncWhitelist{}
	for _, id := range ids {
		out[id] = false
	}

	return out
}

type ItemSyncWhitelist map[blizzardv2.ItemId]bool

func (wl ItemSyncWhitelist) ToItemIds() blizzardv2.ItemIds {
	total := 0
	for _, shouldSync := range wl {
		if !shouldSync {
			continue
		}

		total += 1
	}

	out := make(blizzardv2.ItemIds, total)
	i := 0
	for id, shouldSync := range wl {
		if !shouldSync {
			continue
		}

		out[i] = id

		i += 1
	}

	return out
}
