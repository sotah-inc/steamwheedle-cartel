package run

import (
	"io/ioutil"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta WorkloadState) ComputePricelistHistories(tuple sotah.RegionRealmTimestampTuple) sotah.Message {
	logging.WithField("tuple", tuple).Info("Received tuple")

	// parsing the tuple
	realm := sotah.NewSkeletonRealm(blizzard.RegionName(tuple.RegionName), blizzard.RealmSlug(tuple.RealmSlug))
	targetTime := time.Unix(int64(tuple.TargetTimestamp), 0)

	// gathering auctions from the tuple
	obj, err := sta.auctionsStoreBase.GetFirmObject(realm, targetTime, sta.auctionsBucket)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	reader, err := obj.NewReader(sta.IO.StoreClient.Context)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	aucs, err := blizzard.NewAuctions(data)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	// handling the auctions, target-time, and realm
	logging.WithFields(logrus.Fields{
		"region":        realm.Region.Name,
		"realm":         realm.Slug,
		"last-modified": targetTime.Unix(),
	}).Info("Parsed into live-auctions, handling pricelist-history")
	normalizedTargetTimestamp, err := sta.pricelistHistoriesStoreBase.Handle(
		aucs,
		targetTime,
		realm,
		sta.pricelistHistoriesBucket,
	)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	// deriving a new region-realm-timestamp tuple with the normalized-target-timestamp
	nextTuple := sotah.RegionRealmTimestampTuple{
		TargetTimestamp:  int(normalizedTargetTimestamp),
		RegionRealmTuple: tuple.RegionRealmTuple,
	}
	logging.WithField(
		"normalized-target-timestamp",
		nextTuple.TargetTimestamp,
	).Info("Received normalized-target-timestamp")

	// encoding the resulting tuple
	encodedTuple, err := nextTuple.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	out := sotah.NewMessage()
	out.Data = encodedTuple

	return out
}
