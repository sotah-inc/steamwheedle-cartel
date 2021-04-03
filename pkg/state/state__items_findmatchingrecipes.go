package state

import (
	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta ItemsState) ListenForItemsFindMatchingRecipes(stop ListenStopChan) error {
	return sta.Messenger.Subscribe(
		string(subjects.ItemsFindMatchingRecipes),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			rdMap, err := blizzardv2.NewRecipeIdDescriptionMap(string(natsMsg.Data))
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			irMap, err := sta.ItemsDatabase.FindMatchingItemsFromRecipes(rdMap)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			encodedItemRecipes, err := irMap.EncodeForDelivery()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			// dumping it out
			m.Data = encodedItemRecipes
			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
}
