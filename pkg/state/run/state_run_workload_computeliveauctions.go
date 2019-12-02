package run

import (
	"io/ioutil"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta WorkloadState) ComputeLiveAuctions(tuple sotah.RegionRealmTimestampTuple) sotah.Message {
	realm := sotah.NewSkeletonRealm(blizzard.RegionName(tuple.RegionName), blizzard.RealmSlug(tuple.RealmSlug))
	targetTime := time.Unix(int64(tuple.TargetTimestamp), 0)

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

	logging.WithFields(logrus.Fields{
		"region":        realm.Region.Name,
		"realm":         realm.Slug,
		"last-modified": targetTime.Unix(),
	}).Info("Parsing into live-auctions")
	if err := sta.liveAuctionsStoreBase.Handle(aucs, realm, sta.liveAuctionsBucket); err != nil {
		return sotah.NewErrorMessage(err)
	}

	replyTuple := sotah.RegionRealmSummaryTuple{
		RegionRealmTimestampTuple: tuple,
		ItemIds:                   aucs.ItemIds().ToInts(),
		OwnerNames:                aucs.OwnerNames(),
	}
	encodedReplyTuple, err := replyTuple.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	out := sotah.NewMessage()
	out.Data = encodedReplyTuple

	return out
}
