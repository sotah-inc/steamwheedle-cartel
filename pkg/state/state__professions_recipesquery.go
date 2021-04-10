package state

import (
	"github.com/nats-io/nats.go"
	dCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	ProfessionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta ProfessionsState) ListenForRecipesQuery(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.RecipesQuery), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := ProfessionsDatabase.NewQueryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// querying the items-database
		resp, respCode, err := sta.ProfessionsDatabase.QueryRecipes(request)
		if err != nil {
			logging.WithField(
				"error",
				err.Error(),
			).Error("failed to call sta.ProfessionsDatabase.QueryRecipes")

			m.Err = err.Error()
			m.Code = DatabaseCodeToMessengerCode(respCode)
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}
		if respCode != dCodes.Ok {
			m.Err = "response code was not ok but error was nil"
			m.Code = DatabaseCodeToMessengerCode(respCode)
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// marshalling for messenger
		encodedMessage, err := resp.EncodeForDelivery()
		if err != nil {
			logging.WithField(
				"error",
				err.Error(),
			).Error("failed to call pdBase.GetIdNormalizedNameMap")

			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
