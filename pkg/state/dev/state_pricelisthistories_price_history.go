package dev

import (
	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	dCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
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
