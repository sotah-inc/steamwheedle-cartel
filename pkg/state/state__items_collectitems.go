package state

import (
	"time"

	"github.com/sirupsen/logrus"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta ItemsState) CollectItems(ids blizzardv2.ItemIds) error {
	startTime := time.Now()

	// resolving items to sync
	itemsSyncPayload, err := sta.ItemsDatabase.FilterInItemsToSync(ids)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to filter in items to sync")

		return err
	}

	logging.WithField("items", itemsSyncPayload.Ids).Info("collecting items")

	// starting up workers for resolving items
	itemsOut, err := sta.BlizzardState.ResolveItems(
		sta.RegionsState.RegionComposites.ToList(),
		itemsSyncPayload.Ids,
	)
	if err != nil {
		return err
	}

	// starting up workers for resolving item-medias
	itemMediasIn := make(chan blizzardv2.GetItemMediasInJob)
	itemMediasOut := sta.BlizzardState.ResolveItemMedias(itemMediasIn)

	// starting up an intake queue
	persistItemsIn := make(chan sotah.Item)

	// queueing it all up
	go func() {
		for job := range itemsOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve item")

				continue
			}

			logging.WithField("item-id", job.Id).Info("enqueueing item for item-media resolution")

			itemMediasIn <- blizzardv2.GetItemMediasInJob{Item: job.ItemResponse}
		}

		close(itemMediasIn)
	}()
	go func() {
		for job := range itemMediasOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve item-media")

				continue
			}

			itemIcon, err := job.ItemMediaResponse.GetIcon()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.ItemMediaResponse,
				}).Error("failed to resolve item-icon from item-media")

				continue
			}

			logging.WithField("item-id", job.Item.Id).Info("enqueueing item for persistence")

			persistItemsIn <- sotah.Item{
				BlizzardMeta: job.Item,
				SotahMeta: sotah.ItemMeta{
					LastModified:   sotah.UnixTimestamp(time.Now().Unix()),
					NormalizedName: job.Item.Name,
					ItemIconMeta: sotah.ItemIconMeta{
						URL:        blizzardv2.DefaultGetItemIconURL(itemIcon),
						ObjectName: sotah.NewItemObjectName(sotah.IconName(itemIcon)),
						Icon:       sotah.IconName(itemIcon),
					},
				},
			}
		}

		close(persistItemsIn)
	}()

	totalPersisted, err := sta.ItemsDatabase.PersistItems(persistItemsIn)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist items")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-items")

	return nil
}
