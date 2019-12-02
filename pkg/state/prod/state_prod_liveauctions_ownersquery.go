package prod

import (
	"time"

	"git.sotah.info/steamwheedle-cartel/pkg/database"
	dCodes "git.sotah.info/steamwheedle-cartel/pkg/database/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/logging"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	mCodes "git.sotah.info/steamwheedle-cartel/pkg/messenger/codes"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
	nats "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
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
