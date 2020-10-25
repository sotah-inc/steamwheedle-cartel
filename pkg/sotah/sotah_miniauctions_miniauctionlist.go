package sotah

import (
	"encoding/json"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortdirections"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortkinds"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewMiniAuctionListFromGzipped(gzipEncoded []byte) (MiniAuctionList, error) {
	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return nil, err
	}

	var out MiniAuctionList
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func NewMiniAuctionList(aucs blizzardv2.Auctions) MiniAuctionList {
	ma := NewMiniAuctions(aucs)

	out := make(MiniAuctionList, len(ma))
	i := 0
	for _, mAuction := range ma {
		out[i] = mAuction

		i += 1
	}

	return out
}

type MiniAuctionList []miniAuction

func (maList MiniAuctionList) Limit(count int, page int) (MiniAuctionList, error) {
	alLength := len(maList)
	if alLength == 0 {
		return maList, nil
	}

	start := page * count
	if start > alLength {
		return MiniAuctionList{}, fmt.Errorf("start out of range: %d", start)
	}

	end := start + count
	if end > alLength {
		return maList[start:], nil
	}

	return maList[start:end], nil
}

func (maList MiniAuctionList) Sort(kind sortkinds.SortKind, direction sortdirections.SortDirection) error {
	return newMiniAuctionSorter().sort(kind, direction, maList)
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

func (maList MiniAuctionList) ItemIds() []blizzardv2.ItemId {
	result := map[blizzardv2.ItemId]struct{}{}
	for _, ma := range maList {
		result[ma.ItemId] = struct{}{}
	}

	out := make(blizzardv2.ItemIds, len(result))
	i := 0
	for v := range result {
		out[i] = v

		i += 1
	}

	return out
}

func (maList MiniAuctionList) TotalAuctions() int {
	out := 0
	for _, auc := range maList {
		out += len(auc.AucList)
	}

	return out
}

func (maList MiniAuctionList) TotalQuantity() int {
	out := 0
	for _, auc := range maList {
		out += auc.Quantity * len(auc.AucList)
	}

	return out
}

func (maList MiniAuctionList) TotalBuyout() float64 {
	out := float64(0)
	for _, auc := range maList {
		out += float64(auc.Buyout) * float64(auc.Quantity) * float64(len(auc.AucList))
	}

	return out
}

func (maList MiniAuctionList) AuctionIds() []blizzardv2.AuctionId {
	result := map[blizzardv2.AuctionId]struct{}{}
	for _, mAuction := range maList {
		for _, auc := range mAuction.AucList {
			result[auc] = struct{}{}
		}
	}

	out := make([]blizzardv2.AuctionId, len(result))
	i := 0
	for id := range result {
		out = append(out, id)

		i += 1
	}

	return out
}

func (maList MiniAuctionList) EncodeForStorage() ([]byte, error) {
	jsonEncodedData, err := json.Marshal(maList)
	if err != nil {
		return []byte{}, err
	}

	gzipEncodedData, err := util.GzipEncode(jsonEncodedData)
	if err != nil {
		return []byte{}, err
	}

	return gzipEncodedData, nil
}

func (maList MiniAuctionList) ToItemBuyoutPerListMap() blizzardv2.ItemBuyoutPerSummaryMap {
	result := blizzardv2.NewItemBuyoutPerListMap(maList.ItemIds())
	for _, mAuction := range maList {
		result = result.Insert(mAuction.ItemId, mAuction.BuyoutPer)
	}

	return blizzardv2.NewItemBuyoutPerSummaryMap(result)
}
