package sotah

import (
	"encoding/json"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortdirections"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortkinds"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

// miniauction
func newMiniAuction(auc blizzardv2.Auction) miniAuction {
	var buyoutPer float32
	if auc.Buyout > 0 {
		buyoutPer = float32(auc.Buyout) / float32(auc.Quantity)
	}

	return miniAuction{
		auc.Item.Id,
		auc.Buyout,
		buyoutPer,
		auc.Quantity,
		auc.TimeLeft,
		[]blizzardv2.AuctionId{},
	}
}

type miniAuction struct {
	ItemId    blizzardv2.ItemId      `json:"itemId"`
	Buyout    int64                  `json:"buyout"`
	BuyoutPer float32                `json:"buyoutPer"`
	Quantity  int                    `json:"quantity"`
	TimeLeft  string                 `json:"timeLeft"`
	AucList   []blizzardv2.AuctionId `json:"aucList"`
}

// miniauction-list
func NewMiniAuctionListFromMiniAuctions(ma MiniAuctions) MiniAuctionList {
	out := make(MiniAuctionList, len(ma))
	i := 0
	for _, mAuction := range ma {
		out[i] = mAuction

		i += 1
	}

	return out
}

func NewMiniAuctionListFromGzipped(body []byte) (MiniAuctionList, error) {
	gzipDecodedData, err := util.GzipDecode(body)
	if err != nil {
		return MiniAuctionList{}, err
	}

	return newMiniAuctionList(gzipDecodedData)
}

func newMiniAuctionList(body []byte) (MiniAuctionList, error) {
	maList := &MiniAuctionList{}
	if err := json.Unmarshal(body, maList); err != nil {
		return nil, err
	}

	return *maList, nil
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
	mas := newMiniAuctionSorter()
	return mas.sort(kind, direction, maList)
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

func (maList MiniAuctionList) EncodeForDatabase() ([]byte, error) {
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

// mini-auctions
func NewMiniAuctions(aucs blizzardv2.Auctions) MiniAuctions {
	out := MiniAuctions{}
	for _, auc := range aucs {
		maHash := newMiniAuctionHash(auc)
		if mAuction, ok := out[maHash]; ok {
			mAuction.AucList = append(mAuction.AucList, auc.Id)
			out[maHash] = mAuction

			continue
		}

		mAuction := newMiniAuction(auc)
		mAuction.AucList = append(mAuction.AucList, auc.Id)
		out[maHash] = mAuction
	}

	return out
}

type MiniAuctions map[miniAuctionHash]miniAuction

func newMiniAuctionHash(auc blizzardv2.Auction) miniAuctionHash {
	return miniAuctionHash(fmt.Sprintf(
		"%d-%d-%d-%s",
		auc.Item,
		auc.Buyout,
		auc.Quantity,
		auc.TimeLeft,
	))
}

type miniAuctionHash string
