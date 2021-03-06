package state

import (
	"github.com/nats-io/nats.go"
	LiveAuctionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/liveauctions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta LiveAuctionsState) ListenForPriceList(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.PriceList), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := LiveAuctionsDatabase.NewGetPricelistRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res, dbCode, err := sta.LiveAuctionsDatabases.GetPricelist(req)
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
