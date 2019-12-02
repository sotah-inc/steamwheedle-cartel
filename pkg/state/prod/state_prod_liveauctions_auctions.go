package prod

import (
	"git.sotah.info/steamwheedle-cartel/pkg/database"
	dCodes "git.sotah.info/steamwheedle-cartel/pkg/database/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	mCodes "git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
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
