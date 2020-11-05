package state

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
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

	// gathering profession recipe-ids
	professionRecipeIds, err := sta.ProfessionsDatabase.GetProfessionRecipeIds()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to get profession recipe-ids")

		return err
	}

	// gathering current recipe-ids
	currentRecipeIds, err := sta.ProfessionsDatabase.GetRecipeIds()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get current recipe-ids")

		return err
	}

	// resolving recipe-ids to fetch
	currentRecipeIdsMap := map[blizzardv2.RecipeId]struct{}{}
	for _, id := range currentRecipeIds {
		currentRecipeIdsMap[id] = struct{}{}
	}
	var recipeIdsToFetch []blizzardv2.RecipeId
	for _, id := range professionRecipeIds {
		if _, ok := currentRecipeIdsMap[id]; ok {
			continue
		}

		recipeIdsToFetch = append(recipeIdsToFetch, id)
	}

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
