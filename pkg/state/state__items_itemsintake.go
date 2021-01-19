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

		logging.WithField("items", len(ids)).Info("received")
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
	itemsSyncPayload, err := sta.ItemsDatabase.FilterInItemsToSync(ids)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to filter in items to sync")

		return err
	}

	logging.WithField("items", len(itemsSyncPayload.Ids)).Info("collecting items")

	// starting up an intake queue
	getEncodedItemsOut := sta.LakeClient.GetEncodedItems(itemsSyncPayload.Ids)
	persistItemsIn := make(chan ItemsDatabase.PersistEncodedItemsInJob)

	// queueing it all up
	go func() {
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
