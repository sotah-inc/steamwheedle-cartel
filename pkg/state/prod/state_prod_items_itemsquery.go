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
