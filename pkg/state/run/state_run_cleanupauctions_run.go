package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta CleanupAuctionsState) Handle(regionRealmTuple sotah.RegionRealmTuple) sotah.Message {
	realm := sotah.NewSkeletonRealm(
		blizzard.RegionName(regionRealmTuple.RegionName),
		blizzard.RealmSlug(regionRealmTuple.RealmSlug),
	)

	expiredTimestamps, err := sta.auctionsBase.GetExpiredTimestamps(realm, sta.auctionsBucket)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	deleteResults, err := sta.auctionsBase.DeleteAllFromTimestamps(expiredTimestamps, realm, sta.auctionsBucket)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	res := sotah.CleanupAuctionsPayloadResponse{
		RegionRealmTuple:      sotah.NewRegionRealmTupleFromRealm(realm),
		TotalDeletedCount:     deleteResults.TotalCount,
		TotalDeletedSizeBytes: deleteResults.TotalSize,
	}
	data, err := res.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	m := sotah.NewMessage()
	m.Data = data

	return m
}

func (sta CleanupAuctionsState) Run(data []byte) sotah.Message {
	regionRealmTuple, err := sotah.NewRegionRealmTuple(string(data))
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	return sta.Handle(regionRealmTuple)
}
