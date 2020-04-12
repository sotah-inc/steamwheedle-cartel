package state

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

type ItemBlacklist []blizzardv2.ItemId

func (ib ItemBlacklist) IsPresent(itemId blizzardv2.ItemId) bool {
	for _, blacklistItemId := range ib {
		if blacklistItemId == itemId {
			return true
		}
	}

	return false
}
