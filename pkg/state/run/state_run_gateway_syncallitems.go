package run

import (
	"errors"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta GatewayState) HandleItemIds(ids blizzard.ItemIds) error {
	// batching items together
	logging.WithField("ids", len(ids)).Info("Batching ids together")
	itemIdsBatches := sotah.NewItemIdsBatches(ids, 1000)

	logging.WithField("batches", len(itemIdsBatches)).Info("Handling batches")

	return nil
}

func (sta GatewayState) HandleItemIcons(iconsMap map[string]blizzard.ItemIds) error {
	// batching icons together
	logging.WithField("icons", len(iconsMap)).Info("Batching icons together")
	iconBatches := sotah.NewIconItemsPayloadsBatches(iconsMap, 100)

	// producing messages
	logging.WithField("batches", len(iconBatches)).Info("Handling batches")

	return nil
}

func (sta GatewayState) SyncAllItems(providedItemIds blizzard.ItemIds) error {
	encodedItemIds, err := providedItemIds.EncodeForDelivery()
	if err != nil {
		return err
	}

	startTime := time.Now()

	// filtering in items-to-sync
	response, err := sta.IO.BusClient.Request(sta.filterInItemsToSyncTopic, encodedItemIds, 30*time.Second)
	if err != nil {
		return err
	}

	// optionally halting
	if response.Code != codes.Ok {
		return errors.New("response code was not ok")
	}

	// parsing response data
	syncPayload, err := database.NewItemsSyncPayload(response.Data)
	if err != nil {
		return err
	}

	// handling item-ids
	if len(syncPayload.Ids) == 0 {
		logging.Info("No item-ids in sync-payload, skipping")
	} else {
		if err := sta.HandleItemIds(syncPayload.Ids); err != nil {
			return err
		}
	}

	// handling item-icons
	if len(syncPayload.IconIdsMap) == 0 {
		logging.Info("No item-icons in sync-payload, skipping")
	} else {
		if err := sta.HandleItemIcons(syncPayload.IconIdsMap); err != nil {
			return err
		}
	}

	// reporting metrics
	if err := sta.IO.BusClient.PublishMetrics(metric.Metrics{
		"sync_all_items_duration": int(int64(time.Since(startTime)) / 1000 / 1000 / 1000),
		"sync_all_items_ids":      len(syncPayload.Ids),
		"sync_all_items_icons":    len(syncPayload.IconIdsMap),
	}); err != nil {
		return err
	}

	return nil
}
