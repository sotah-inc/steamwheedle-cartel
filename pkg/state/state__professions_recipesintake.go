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

func (sta ProfessionsState) ListenForRecipesIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.RecipesIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		if err := sta.RecipesIntake(); err != nil {
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

func (sta ProfessionsState) RecipesIntake() error {
	startTime := time.Now()

	// resolving profession recipe-ids to check
	professionRecipeIds, err := sta.ProfessionsDatabase.GetProfessionRecipeIds()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to get profession recipe-ids")

		return err
	}

	// resolving current recipe-ids
	currentRecipeIds, err := sta.ProfessionsDatabase.GetRecipeIds()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get current recipe-ids")

		return err
	}

	// resolving recipe-ids to fetch
	recipeIdsToFetch := blizzardv2.NewRecipeIdsFromMap(
		blizzardv2.NewRecipeIdMap(professionRecipeIds).Exclude(currentRecipeIds),
	)

	logging.WithFields(logrus.Fields{
		"profession-recipe-ids": len(professionRecipeIds),
		"current-recipe-ids":    len(currentRecipeIds),
		"recipe-ids-to-fetch":   len(recipeIdsToFetch),
	}).Info("collecting recipe-ids")

	logging.WithFields(logrus.Fields{
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in skill-tier-intake")

	return nil
}
