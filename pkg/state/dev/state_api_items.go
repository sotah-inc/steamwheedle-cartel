package dev

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"

	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta *APIState) ListenForItems(stop state.ListenStopChan) error {
	err := sta.messenger.Subscribe(string(subjects.Items), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		iRequest, err := state.NewItemsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		iMap, err := sta.itemsDatabase.FindItems(iRequest.ItemIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		iResponse := state.ItemsResponse{Items: iMap}
		data, err := iResponse.EncodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		sta.messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
