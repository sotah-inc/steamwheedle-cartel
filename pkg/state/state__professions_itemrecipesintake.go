package state

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (sta ProfessionsState) ListenForItemRecipesIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemRecipesIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		irMap, err := blizzardv2.NewItemRecipesMap(string(natsMsg.Data))
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		if err := sta.ItemRecipesIntake(irMap); err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta ProfessionsState) ItemRecipesIntake(irMap blizzardv2.ItemRecipesMap) error {
	startTime := time.Now()

	if err := sta.ProfessionsDatabase.PersistItemRecipes(irMap); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist item-recipes")

		return err
	}

	logging.WithFields(logrus.Fields{
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("persisted item-recipes")

	return nil
}
