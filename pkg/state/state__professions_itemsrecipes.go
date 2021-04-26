package state

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewItemsRecipesRequest(body []byte) (ItemsRecipesRequest, error) {
	out := ItemsRecipesRequest{}

	if err := json.Unmarshal(body, &out); err != nil {
		return ItemsRecipesRequest{}, err
	}

	return out, nil
}

type ItemsRecipesRequest struct {
	ItemIds blizzardv2.ItemIds `json:"item_ids"`
}

func (sta ProfessionsState) ListenForItemsRecipes(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemsRecipes), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := NewItemsRecipesRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// gathering items-recipes map
		irMap, err := sta.ProfessionsDatabase.GetItemRecipesMap(request.ItemIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithFields(logrus.Fields{
			"items":        request.ItemIds,
			"item-recipes": irMap,
		}).Info("resolved item-recipes with items")

		// marshalling for messenger
		encodedMessage, err := irMap.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = encodedMessage
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
