package fn

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func (sta ComputePricelistHistoriesState) Handle(job bus.LoadRegionRealmTimestampsInJob) bus.Message {
	m := bus.NewMessage()

	realm, targetTime := job.ToRealmTime()

	obj, err := sta.auctionsStoreBase.GetFirmObject(realm, targetTime, sta.auctionsBucket)
	if err != nil {
		m.Err = err.Error()
		m.Code = codes.GenericError

		return m
	}

	reader, err := obj.NewReader(sta.IO.StoreClient.Context)
	if err != nil {
		m.Err = err.Error()
		m.Code = codes.GenericError

		return m
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		m.Err = err.Error()
		m.Code = codes.GenericError

		return m
	}

	aucs, err := blizzard.NewAuctions(data)
	if err != nil {
		m.Err = err.Error()
		m.Code = codes.GenericError

		return m
	}

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
		m.Err = err.Error()
		m.Code = codes.GenericError

		return m
	}

	logging.WithField("normalized-target-timestamp", normalizedTargetTimestamp).Info("e")

	replyTuple := bus.RegionRealmTimestampTuple{
		RegionName:                job.RegionName,
		RealmSlug:                 job.RealmSlug,
		NormalizedTargetTimestamp: int(normalizedTargetTimestamp),
	}
	encodedReplyRequest, err := replyTuple.EncodeForDelivery()
	if err != nil {
		m.Err = err.Error()
		m.Code = codes.GenericError

		return m
	}
	m.Data = encodedReplyRequest

	return m
}

func (sta ComputePricelistHistoriesState) Run(data string) error {
	var in bus.Message
	if err := json.Unmarshal([]byte(data), &in); err != nil {
		return err
	}

	job, err := bus.NewLoadRegionRealmTimestampsInJob(in.Data)
	if err != nil {
		return err
	}

	msg := sta.Handle(job)
	msg.ReplyToId = in.ReplyToId
	if _, err := sta.IO.BusClient.ReplyTo(in, msg); err != nil {
		return err
	}

	return nil
}
