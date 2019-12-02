package prod

import (
	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/state/subjects"
)

func (itemsState ItemsState) ListenForItems(stop state.ListenStopChan) error {
	err := itemsState.IO.Messenger.Subscribe(string(subjects.Items), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		iRequest, err := state.NewItemsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			itemsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		iMap, err := itemsState.IO.Databases.ItemsDatabase.FindItems(iRequest.ItemIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			itemsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		iResponse := state.ItemsResponse{Items: iMap}
		data, err := iResponse.EncodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			itemsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		itemsState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
