package state

import (
	"github.com/nats-io/nats.go"
	LiveAuctionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/liveauctions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta LiveAuctionsState) ListenForAuctions(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Auctions), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		qr, err := LiveAuctionsDatabase.NewQueryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res, dbCode, err := sta.LiveAuctionsDatabases.QueryAuctions(qr)
		if err != nil {
			m.Err = err.Error()
			m.Code = DatabaseCodeToMessengerCode(dbCode)
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		data, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
