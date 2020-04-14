package state

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (sta DiskAuctionsState) CollectItems(ids blizzardv2.ItemIds) error {
	// resolving items to sync
	itemsSyncPayload, err := sta.ItemsDatabase.FilterInItemsToSync(ids)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to filter in items to sync")

		return err
	}

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
				logging.WithField("error", err.Error()).Error("failed to resolve item-icon from item-media")

				continue
			}

			persistItemsIn <- sotah.Item{
				BlizzardMeta: job.Item,
				SotahMeta: sotah.ItemMeta{
					LastModified:   sotah.UnixTime{Time: time.Now()},
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

	return sta.ItemsDatabase.PersistItems(persistItemsIn)
}
