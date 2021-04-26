package state

import (
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
		if err := sta.itemsIntake(ids); err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta ItemsState) itemsIntake(ids blizzardv2.ItemIds) error {
	startTime := time.Now()

	// resolving items to sync
	itemIds, err := sta.ItemsDatabase.FilterInItemsToSync(ids)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to filter in items to sync")

		return err
	}

	if len(itemIds) == 0 {
		logging.Info("skipping items-intake as none were filtered in")

		return nil
	}

	logging.WithField("items", len(itemIds)).Info("collecting items")

	// starting up an intake queue
	getEncodedItemsOut, erroneousItemIdsOut := sta.LakeClient.GetEncodedItems(itemIds)
	persistItemsIn := make(chan ItemsDatabase.PersistEncodedItemsInJob)
	itemClassItemsOut := make(chan blizzardv2.ItemClassItemsMap)

	// queueing it all up
	go func() {
		itemClassItems := blizzardv2.ItemClassItemsMap{}
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
		}

		close(persistItemsIn)

		itemClassItemsOut <- itemClassItems
		close(itemClassItemsOut)
	}()

	totalPersisted, err := sta.ItemsDatabase.PersistEncodedItems(persistItemsIn)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist items")

		return err
	}

	erroneousItemIds := <-erroneousItemIdsOut
	if err := sta.ItemsDatabase.PersistBlacklistedIds(erroneousItemIds); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist blacklisted item-ids")

		return err
	}

	itemClassItems := <-itemClassItemsOut

	logging.WithField("item-class-items", itemClassItems).Info("persisting item-class items")

	if err := sta.ItemsDatabase.ReceiveItemClassItemsMap(itemClassItems); err != nil {
		logging.WithField("error", err.Error()).Error("failed to receive item-class-items")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-items")

	return nil
}
