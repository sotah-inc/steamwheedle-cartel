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

func (itemsState ItemsState) ListenForItemsQuery(stop state.ListenStopChan) error {
	err := itemsState.IO.Messenger.Subscribe(string(subjects.ItemsQuery), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := database.NewQueryItemsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			itemsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// querying the items-database
		startTime := time.Now()
		resp, respCode, err := itemsState.IO.Databases.ItemsDatabase.QueryItems(request)
		if err != nil {
			m.Err = err.Error()
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			itemsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}
		if respCode != dCodes.Ok {
			m.Err = "response code was not ok but error was nil"
			m.Code = state.DatabaseCodeToMessengerCode(respCode)
			itemsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		duration := time.Since(startTime)
		logging.WithFields(logrus.Fields{
			"query":          request.Query,
			"duration-in-ms": int64(duration) / 1000 / 1000,
		}).Info("Queried items")

		// marshalling for messenger
		encodedMessage, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			itemsState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		itemsState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
