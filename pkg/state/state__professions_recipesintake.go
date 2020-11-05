package state

import (
	"time"

	ProfessionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions"

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

	// starting up an intake queue
	getEncodedRecipesOut := sta.LakeClient.GetEncodedRecipes(recipeIdsToFetch)
	persistRecipesIn := make(chan ProfessionsDatabase.PersistEncodedRecipesInJob)

	// queueing it all up
	go func() {
		for job := range getEncodedRecipesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve recipe")

				continue
			}

			logging.WithField("recipe-id", job.Id()).Info("enqueueing recipe for persistence")

			persistRecipesIn <- ProfessionsDatabase.PersistEncodedRecipesInJob{
				RecipeId:      job.Id(),
				EncodedRecipe: job.EncodedRecipe(),
			}
		}

		close(persistRecipesIn)
	}()

	totalPersisted, err := sta.ProfessionsDatabase.PersistEncodedRecipes(persistRecipesIn)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist recipe")

		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in recipes-intake")

	return nil
}
