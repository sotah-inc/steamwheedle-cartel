package prod

import (
	nats "github.com/nats-io/go-nats"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/database"
	dCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/database/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/state/subjects"
)

func (liveAuctionsState ProdLiveAuctionsState) ListenForAuctions(stop state.ListenStopChan) error {
	err := liveAuctionsState.IO.Messenger.Subscribe(string(subjects.Auctions), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		qRequest, err := database.NewQueryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		qResponse, respCode, err := liveAuctionsState.IO.Databases.LiveAuctionsDatabases.QueryAuctions(qRequest)
		if err != nil {
			m.Err = err.Error()
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}
		if respCode != dCodes.Ok {
			m.Err = "response code was not ok but error was nil"
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// encoding the auctions list for output
		data, err := qResponse.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
