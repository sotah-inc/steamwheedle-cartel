package state

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewItemSubjectsByItemClassRequest(data []byte) (ItemSubjectsByItemClassRequest, error) {
	out := ItemSubjectsByItemClassRequest{}

	if err := json.Unmarshal(data, &out); err != nil {
		return ItemSubjectsByItemClassRequest{}, err
	}

	return out, nil
}

type ItemSubjectsByItemClassRequest struct {
	ItemClassId itemclass.Id            `json:"item_class_id"`
	Version     gameversion.GameVersion `json:"game_version"`
}

func (req ItemSubjectsByItemClassRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}

func (sta ItemsState) ListenForItemSubjectsByItemClass(stop ListenStopChan) error {
	return sta.Messenger.Subscribe(
		string(subjects.ItemSubjectsByItemClass),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			logging.Info("handling request for items item-subjects-by-item-class")

			req, err := NewItemSubjectsByItemClassRequest(natsMsg.Data)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			itemIds, err := sta.ItemsDatabase.GetItemClassItemIds(req.ItemClassId)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			itemSubjects, err := sta.ItemsDatabase.GetItemSubjects(req.Version, itemIds)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

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
