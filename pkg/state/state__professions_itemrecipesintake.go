package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2" // nolint:lll
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

	logging.WithFields(logrus.Fields{
		"item-recipes": len(irMap),
	}).Info("handling request for professions item-recipes intake")

	// resolving existing ir-map and merging results in
	currentIrMap, err := sta.ProfessionsDatabase.GetItemRecipesMap(irMap.ItemIds())
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to resolve item-recipes map")

		return err
	}

	logging.WithFields(logrus.Fields{
		"item-recipes":         len(irMap),
		"current-item-recipes": len(currentIrMap),
	}).Info("found current item-recipes")

	nextIrMap := currentIrMap.Merge(irMap)

	logging.WithFields(logrus.Fields{
		"item-recipes":         len(irMap),
		"current-item-recipes": len(currentIrMap),
		"merged-item-recipes":  len(nextIrMap),
	}).Info("resolved merged item-recipes")

	// pushing next ir-map out
	if err := sta.ProfessionsDatabase.PersistItemRecipes(nextIrMap); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist item-recipes")

		return err
	}

	logging.WithFields(logrus.Fields{
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("persisted item-recipes")

	return nil
}
