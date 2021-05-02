package state

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	ItemsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/items" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewItemsIntakeResponse(jsonEncoded []byte) (ItemsIntakeResponse, error) {
	out := ItemsIntakeResponse{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return ItemsIntakeResponse{}, err
	}

	return out, nil
}

type ItemsIntakeResponse struct {
	TotalPersisted int `json:"total_persisted"`
}

func (resp ItemsIntakeResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(resp)
}

func (sta ItemsState) ListenForItemsIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemsIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		ids, err := blizzardv2.NewItemIds(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("items", len(ids)).Info("received item-ids")
		resp, err := sta.itemsIntake(ids)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		encodedResponse, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(encodedResponse)

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta ItemsState) itemsIntake(ids blizzardv2.ItemIds) (ItemsIntakeResponse, error) {
	startTime := time.Now()

	// resolving items to sync
	itemIds, err := sta.ItemsDatabase.FilterInItemsToSync(ids)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to filter in items to sync")

		return ItemsIntakeResponse{}, err
	}

	if len(itemIds) == 0 {
		logging.Info("skipping items-intake as none were filtered in")

		return ItemsIntakeResponse{TotalPersisted: 0}, nil
	}

	logging.WithField("items", len(itemIds)).Info("collecting items")

	// starting up an intake queue
	getEncodedItemsOut, erroneousItemIdsOut := sta.LakeClient.GetEncodedItems(itemIds)
	persistItemsIn := make(chan ItemsDatabase.PersistEncodedItemsInJob)
	itemClassItemsOut := make(chan blizzardv2.ItemClassItemsMap)
	itemVendorPricesOut := make(chan map[blizzardv2.ItemId]blizzardv2.PriceValue)

	// queueing it all up
	go func() {
		itemClassItems := blizzardv2.ItemClassItemsMap{}
		itemVendorPrices := map[blizzardv2.ItemId]blizzardv2.PriceValue{}
		for job := range getEncodedItemsOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve item")

				continue
			}

			logging.WithField("item-id", job.Id()).Info("enqueueing item for persistence")

			persistItemsIn <- ItemsDatabase.PersistEncodedItemsInJob{
				Id:                    job.Id(),
				EncodedItem:           job.EncodedItem(),
				EncodedNormalizedName: job.EncodedNormalizedName(),
			}

			itemClassItems = itemClassItems.Insert(job.ItemClass(), job.Id())

			if job.IsVendorItem() {
				itemVendorPrices[job.Id()] = job.VendorPrice()
			}
		}

		close(persistItemsIn)

		itemClassItemsOut <- itemClassItems
		close(itemClassItemsOut)

		itemVendorPricesOut <- itemVendorPrices
		close(itemVendorPricesOut)
	}()

	totalPersisted, err := sta.ItemsDatabase.PersistEncodedItems(persistItemsIn)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist items")

		return ItemsIntakeResponse{}, err
	}

	erroneousItemIds := <-erroneousItemIdsOut
	if err := sta.ItemsDatabase.PersistBlacklistedIds(erroneousItemIds); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist blacklisted item-ids")

		return ItemsIntakeResponse{}, err
	}

	itemClassItems := <-itemClassItemsOut
	if err := sta.ItemsDatabase.ReceiveItemClassItemsMap(itemClassItems); err != nil {
		logging.WithField("error", err.Error()).Error("failed to receive item-class-items")

		return ItemsIntakeResponse{}, err
	}

	itemVendorPrices := <-itemVendorPricesOut
	if len(itemVendorPrices) > 0 {
		logging.WithField("item-vendor-prices", itemVendorPrices).Info("received item-vendor-prices")
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-items")

	return ItemsIntakeResponse{TotalPersisted: totalPersisted}, nil
}
