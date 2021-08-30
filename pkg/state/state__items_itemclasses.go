package state

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ItemsClassesResponse struct {
	ItemClasses []blizzardv2.ItemClassResponse `json:"item_classes"`
}

func (sta ItemsState) ListenForItemClasses(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemClasses), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		itemClasses, err := sta.ItemsDatabase.GetItemClasses()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		jsonEncoded, err := json.Marshal(ItemsClassesResponse{ItemClasses: itemClasses})
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithFields(logrus.Fields{
			"subject": subjects.ItemClasses,
			"message": string(jsonEncoded),
		}).Info("sending on subject")

		m.Data = string(jsonEncoded)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
