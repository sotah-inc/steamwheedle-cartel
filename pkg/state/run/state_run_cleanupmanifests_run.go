package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta CleanupManifestsState) Handle(regionRealmTuple sotah.RegionRealmTuple) sotah.Message {
	realm := sotah.NewSkeletonRealm(
		blizzard.RegionName(regionRealmTuple.RegionName),
		blizzard.RealmSlug(regionRealmTuple.RealmSlug),
	)

	expiredTimestamps, err := sta.manifestBase.GetExpiredTimestamps(realm, sta.manifestBucket)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	totalDeleted, err := sta.manifestBase.DeleteAllFromTimestamps(expiredTimestamps, realm, sta.manifestBucket)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	res := sotah.CleanupManifestPayloadResponse{
		RegionRealmTuple: sotah.NewRegionRealmTupleFromRealm(realm),
		TotalDeleted:     totalDeleted,
	}
	data, err := res.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	m := sotah.NewMessage()
	m.Data = data

	return m
}

func (sta CleanupManifestsState) Run(data []byte) sotah.Message {
	regionRealmTuple, err := sotah.NewRegionRealmTuple(string(data))
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	return sta.Handle(regionRealmTuple)
}
