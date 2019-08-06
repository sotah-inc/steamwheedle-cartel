package run

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta GatewayState) HandleItemIds(ids blizzard.ItemIds) error { // generating new act client
	logging.WithField("endpoint-url", sta.actEndpoints.SyncItems).Info("Producing act client for sync-items act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.SyncItems)
	if err != nil {
		return err
	}

	// batching items together
	logging.WithField("ids", len(ids)).Info("Batching ids together")
	itemIdsBatches := sotah.NewItemIdsBatches(ids, 1000)

	// calling act client with item-ids batches
	logging.Info("Calling sync-items with act client")
	for outJob := range actClient.SyncItems(itemIdsBatches) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to sync items")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusCreated:
			continue
		default:
			logging.WithFields(logrus.Fields{
				"status-code": outJob.Data.Code,
				"data":        fmt.Sprintf("%.25s", string(outJob.Data.Body)),
			}).Error("Response code for act call was invalid")
		}
	}

	return nil
}

func (sta GatewayState) HandleItemIcons(iconsMap map[string]blizzard.ItemIds) error {
	logging.WithField(
		"endpoint-url",
		sta.actEndpoints.SyncItemIcons,
	).Info("Producing act client for sync-item-icons act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.SyncItemIcons)
	if err != nil {
		return err
	}

	// batching icons together
	logging.WithField("icons", len(iconsMap)).Info("Batching icons together")
	iconBatches := sotah.NewIconItemsPayloadsBatches(iconsMap, 100)

	// calling act client with item-icons batches
	logging.Info("Calling item-icons with act client")
	for outJob := range actClient.SyncItemIcons(iconBatches) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to sync item-icons")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusCreated:
			continue
		default:
			logging.WithFields(logrus.Fields{
				"status-code": outJob.Data.Code,
				"data":        fmt.Sprintf("%.25s", string(outJob.Data.Body)),
			}).Error("Response code for act call was invalid")
		}
	}

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
