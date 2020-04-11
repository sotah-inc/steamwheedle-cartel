package sotah

import (
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func NewMiniAuctions(aucs blizzardv2.Auctions) MiniAuctions {
	out := MiniAuctions{}
	for _, auc := range aucs {
		out = out.Insert(auc)
	}

	return out
}

type MiniAuctions map[miniAuctionHash]miniAuction

func (mAuctions MiniAuctions) Insert(auc blizzardv2.Auction) MiniAuctions {
	maHash := newMiniAuctionHash(auc)

	mAuction := func() miniAuction {
		if found, ok := mAuctions[maHash]; ok {
			return found
		}

		return newMiniAuction(auc)
	}()

	mAuction.AucList = append(mAuction.AucList, auc.Id)
	mAuctions[maHash] = mAuction

	return mAuctions
}

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
