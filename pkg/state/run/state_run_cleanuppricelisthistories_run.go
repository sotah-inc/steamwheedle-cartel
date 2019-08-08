package run

import (
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta CleanupPricelistHistoriesState) Run(tuple sotah.RegionRealmTuple) sotah.Message {
	logging.WithFields(logrus.Fields{"tuple": tuple}).Info("Handling")

	realm := sotah.NewSkeletonRealm(blizzard.RegionName(tuple.RegionName), blizzard.RealmSlug(tuple.RealmSlug))
	expiredTimestamps, err := sta.pricelistHistoriesBase.GetExpiredTimestamps(realm, sta.pricelistHistoriesBucket)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	totalDeleted, err := sta.pricelistHistoriesBase.DeleteAll(realm, expiredTimestamps, sta.pricelistHistoriesBucket)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	res := sotah.CleanupPricelistPayloadResponse{
		RegionName:   tuple.RegionName,
		RealmSlug:    tuple.RealmSlug,
		TotalDeleted: totalDeleted,
	}
	data, err := res.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	out := sotah.NewMessage()
	out.Data = data

	return out
}
