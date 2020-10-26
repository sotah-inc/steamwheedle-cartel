package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

type MiniAuctionListFilterCriteria struct {
	ItemIds blizzardv2.ItemIds
	PetIds  blizzardv2.PetIds
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

func (mafList MiniAuctionFlaggedList) ToMiniAuctionList() MiniAuctionList {
	out := make(MiniAuctionList, len(mafList))
	for i, maf := range mafList {
		out[i] = maf.auc
	}

	return out
}

func (mafList MiniAuctionFlaggedList) Flag(criteria MiniAuctionListFilterCriteria) MiniAuctionFlaggedList {
	petIdsMap := criteria.PetIds.ToUniqueMap()
	itemIdsMap := criteria.ItemIds.ToUniqueMap()

	out := make(MiniAuctionFlaggedList, len(mafList))
	for i, maf := range mafList {
		out[i] = MiniAuctionFlagged{
			auc:     maf.auc,
			flagged: petIdsMap.Exists(maf.auc.PetSpeciesId) || itemIdsMap.Exists(maf.auc.ItemId),
		}
	}

	return out
}

func (mafList MiniAuctionFlaggedList) FilterInFlagged() MiniAuctionFlaggedList {
	out := MiniAuctionFlaggedList{}
	for _, maf := range mafList {
		if !maf.flagged {
			continue
		}

		out = append(out, maf)
	}

	return out
}

type MiniAuctionFlagged struct {
	auc     miniAuction
	flagged bool
}
