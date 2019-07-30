package run

import (
	"errors"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta SyncItemsState) PublishToReceiveSyncedItems(results sotah.ItemIdNameMap) sotah.Message {
	data, err := results.EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	replyMessage, err := sta.IO.BusClient.Request(sta.receiveSyncedItemsTopic, data, 10*time.Second)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	if replyMessage.Code != codes.Ok {
		return sotah.NewErrorMessage(errors.New("reply code was not ok"))
	}

	return sotah.NewMessage()
}

func (sta SyncItemsState) Handle(ids blizzard.ItemIds) (sotah.ItemIdNameMap, error) {
	return sotah.ItemIdNameMap{}, nil
}

func (sta SyncItemsState) Run(ids blizzard.ItemIds) sotah.Message {
	results, err := sta.Handle(ids)
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	return sta.PublishToReceiveSyncedItems(results)
}
