package dev

import (
	"git.sotah.info/steamwheedle-cartel/pkg/database"
	dCodes "git.sotah.info/steamwheedle-cartel/pkg/database/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	mCodes "git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
)

func (sta PricelistHistoriesState) ListenForPriceListHistory(stop state.ListenStopChan) error {
	err := sta.IO.Messenger.Subscribe(string(subjects.PriceListHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := database.NewGetPricelistHistoryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// querying the live-auctions-databases
		resp, respCode, err := sta.IO.Databases.PricelistHistoryDatabases.GetPricelistHistory(request)
		if err != nil {
			m.Err = err.Error()
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}
		if respCode != dCodes.Ok {
			m.Err = "response code was not ok but error was nil"
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// marshalling for messenger
		encodedMessage, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = encodedMessage
		sta.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
