package state

import (
	"encoding/base64"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
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

			jsonEncoded, err := irMap.FilterBlank().EncodeForDelivery()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			gzipEncoded, err := util.GzipEncode(jsonEncoded)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			// dumping it out
			m.Data = base64.StdEncoding.EncodeToString(gzipEncoded)
			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
}
