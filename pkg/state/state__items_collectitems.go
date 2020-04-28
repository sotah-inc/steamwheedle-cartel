package state

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
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

	logging.WithField("items", len(itemsSyncPayload.Ids)).Info("collecting items")

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
	persistItemsIn := make(chan database.PersistEncodedItemsInJob)

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

			normalizedName, err := func() (locale.Mapping, error) {
				foundName := job.Item.Name
				if _, ok := foundName[locale.EnUS]; !ok {
					return locale.Mapping{}, errors.New("failed to resolve enUS name")
				}

				foundName[locale.EnUS], err = sotah.NormalizeString(foundName[locale.EnUS])
				if err != nil {
					return locale.Mapping{}, err
				}

				return foundName, nil
			}()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.ItemMediaResponse,
				}).Error("failed to normalize name")

				continue
			}

			item := sotah.Item{
				BlizzardMeta: job.Item,
				SotahMeta: sotah.ItemMeta{
					LastModified:   sotah.UnixTimestamp(time.Now().Unix()),
					NormalizedName: normalizedName,
					ItemIconMeta: sotah.ItemIconMeta{
						URL:        blizzardv2.DefaultGetItemIconURL(itemIcon),
						ObjectName: sotah.NewItemObjectName(sotah.IconName(itemIcon)),
						Icon:       sotah.IconName(itemIcon),
					},
				},
			}

			encodedItem, err := item.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"item":  item.BlizzardMeta.Id,
				}).Error("failed to encode item for storage")

				continue
			}

			encodedNormalizedName, err := normalizedName.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error": err.Error(),
					"item":  item.BlizzardMeta.Id,
				}).Error("failed to encode normalized-name for storage")

				continue
			}

			persistItemsIn <- database.PersistEncodedItemsInJob{
				Id:                    job.Item.Id,
				EncodedItem:           encodedItem,
				EncodedNormalizedName: encodedNormalizedName,
			}
		}

		close(persistItemsIn)
	}()

	totalPersisted, err := sta.ItemsDatabase.PersistEncodedItems(persistItemsIn)
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
