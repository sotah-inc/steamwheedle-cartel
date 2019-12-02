package dev

import (
	"git.sotah.info/steamwheedle-cartel/pkg/state"

	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
)

func (sta APIState) ListenForItems(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.Items), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		iRequest, err := state.NewItemsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		iMap, err := iRequest.Resolve(sta.State)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		iResponse := state.ItemsResponse{Items: iMap}
		data, err := iResponse.EncodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		sta.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
