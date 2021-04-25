package state

import (
	"strconv"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta ItemsState) ListenForItemSubjectsByItemClass(stop ListenStopChan) error {
	return sta.Messenger.Subscribe(
		string(subjects.ItemSubjectsByItemClass),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			logging.Info("handling request for items item-subjects-by-item-class")

			itemClassId, err := strconv.Atoi(string(natsMsg.Data))
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			logging.WithField("item-class-id", itemClassId).Info("using item-class id")

			itemIds, err := sta.ItemsDatabase.GetItemClassItemIds(itemclass.Id(itemClassId))
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			logging.WithField("item-ids", itemIds).Info("found item-ids ")

			itemSubjects, err := sta.ItemsDatabase.GetItemSubjects(itemIds)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			logging.WithField("item-subjects", itemSubjects).Info("found item-subjects")

			encodedItemSubjects, err := itemSubjects.EncodeForDelivery()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			// dumping it out
			m.Data = encodedItemSubjects
			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
}
