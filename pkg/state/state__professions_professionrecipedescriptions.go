package state

import (
	"strconv"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta ProfessionsState) ListenForProfessionRecipeSubjects(stop ListenStopChan) error {
	return sta.Messenger.Subscribe(
		string(subjects.ProfessionRecipeSubjects),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			professionId, err := strconv.Atoi(string(natsMsg.Data))
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			recipeIds, err := sta.ProfessionsDatabase.GetRecipeIdsByProfessionId(
				blizzardv2.ProfessionId(professionId),
			)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			rsMap, err := sta.ProfessionsDatabase.GetRecipeSubjects(recipeIds)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			// marshalling for messenger
			encodedMessage, err := rsMap.EncodeForDelivery()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			// dumping it out
			m.Data = encodedMessage
			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
}
