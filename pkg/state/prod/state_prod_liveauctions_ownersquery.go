package prod

import (
	"time"

	nats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	dCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (liveAuctionsState ProdLiveAuctionsState) ListenForOwnersQuery(stop state.ListenStopChan) error {
	err := liveAuctionsState.IO.Messenger.Subscribe(string(subjects.OwnersQuery), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := database.NewQueryOwnersRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// querying the live-auctions-databases
		startTime := time.Now()
		resp, respCode, err := liveAuctionsState.IO.Databases.LiveAuctionsDatabases.QueryOwners(request)
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

		duration := time.Since(startTime)
		logging.WithFields(logrus.Fields{
			"region":         request.RegionName,
			"realm":          request.RealmSlug,
			"query":          request.Query,
			"duration-in-ms": int64(duration) / 1000 / 1000,
		}).Info("Queried owners")

		// marshalling for messenger
		encodedMessage, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		liveAuctionsState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
