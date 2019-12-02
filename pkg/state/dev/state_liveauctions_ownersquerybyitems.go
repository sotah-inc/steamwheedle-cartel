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

func (laState LiveAuctionsState) ListenForOwnersQueryByItems(stop state.ListenStopChan) error {
	err := laState.IO.Messenger.Subscribe(string(subjects.OwnersQueryByItems), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := database.NewQueryOwnersByItemsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// querying the live-auctions-databases
		resp, respCode, err := laState.IO.Databases.LiveAuctionsDatabases.QueryOwnersByItems(request)
		if err != nil {
			m.Err = err.Error()
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}
		if respCode != dCodes.Ok {
			m.Err = "response code was not ok but error was nil"
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// marshalling for messenger
		encodedMessage, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = encodedMessage
		laState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
