package run

import (
	"io/ioutil"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func (sta ComputeLiveAuctionsState) Handle(tuple sotah.RegionRealmTimestampTuple) sotah.Message {
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

func (sta ComputeLiveAuctionsState) Run(data []byte) sotah.Message {
	tuple, err := sotah.NewRegionRealmTimestampTuple(string(data))
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	return sta.Handle(tuple)
}
