package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type MiniAuctionListFilterCriteria struct {
	ItemIds []blizzardv2.ItemId
	PetIds  []blizzardv2.PetId
}

func (c MiniAuctionListFilterCriteria) IsEmpty() bool {
	return len(c.ItemIds) == 0 && len(c.PetIds) == 0
}

func NewMiniAuctionFlaggedList(input MiniAuctionList) MiniAuctionFlaggedList {
	out := make(MiniAuctionFlaggedList, len(input))
	for i, auc := range input {
		out[i] = MiniAuctionFlagged{
			auc:     auc,
			flagged: false,
		}
	}

	return out
}

type MiniAuctionFlaggedList []MiniAuctionFlagged

type MiniAuctionFlagged struct {
	auc     miniAuction
	flagged bool
}

func (maList MiniAuctionList) FilterByItemIds(itemIds []blizzardv2.ItemId) MiniAuctionList {
	out := MiniAuctionList{}
	for _, ma := range maList {
		for _, id := range itemIds {
			if ma.ItemId == id {
				out = append(out, ma)
			}
		}
	}

	return out
}

func (maList MiniAuctionList) FilterByPetIds(petIds []blizzardv2.PetId) MiniAuctionList {
	out := MiniAuctionList{}
	for _, ma := range maList {
		for _, id := range petIds {
			if ma.PetSpeciesId == id {
				out = append(out, ma)
			}
		}
	}

	return out
}
