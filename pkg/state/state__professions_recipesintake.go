package state

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	ProfessionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions/professionsflags"    // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func (sta ProfessionsState) ListenForRecipesIntake(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.RecipesIntake), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		resp, err := sta.RecipesIntake()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		encodedResp, err := resp.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = encodedResp

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func NewRecipesIntakeResponse(base64Encoded string) (RecipesIntakeResponse, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return RecipesIntakeResponse{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RecipesIntakeResponse{}, err
	}

	out := RecipesIntakeResponse{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return RecipesIntakeResponse{}, err
	}

	return out, nil
}

type RecipesIntakeResponse struct {
	RecipeItemIds []blizzardv2.ItemId `json:"recipe_item_ids"`
}

func (resp RecipesIntakeResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (sta ProfessionsState) RecipesIntake() (RecipesIntakeResponse, error) {
	isComplete, err := sta.ProfessionsDatabase.IsComplete(string(professionsflags.Recipes))
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to get is-complete for professions-database")

		return RecipesIntakeResponse{}, err
	}

	if isComplete {
		return RecipesIntakeResponse{}, nil
	}

	startTime := time.Now()

	// gathering profession recipes-group
	recipesGroup, err := sta.ProfessionsDatabase.GetRecipesGroup()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to get profession recipes-group")

		return RecipesIntakeResponse{}, err
	}

	// gathering current recipe-ids
	currentRecipeIds, err := sta.ProfessionsDatabase.GetRecipeIds()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to get current recipe-ids")

		return RecipesIntakeResponse{}, err
	}

	// resolving recipe-ids to fetch
	recipesGroupToFetch := recipesGroup.FilterOut(currentRecipeIds)
	totalRecipesToFetch := recipesGroupToFetch.TotalRecipes()
	if totalRecipesToFetch == 0 {
		return RecipesIntakeResponse{RecipeItemIds: []blizzardv2.ItemId{}}, nil
	}

	logging.WithFields(logrus.Fields{
		"profession-recipe-ids": recipesGroup.TotalRecipes(),
		"current-recipe-ids":    len(currentRecipeIds),
		"recipe-ids-to-fetch":   totalRecipesToFetch,
	}).Info("collecting recipe-ids")

	// starting up an intake queue
	getEncodedRecipesOut := sta.LakeClient.GetEncodedRecipes(recipesGroupToFetch)
	persistRecipesIn := make(chan ProfessionsDatabase.PersistEncodedRecipesInJob)
	itemRecipesOut := make(chan blizzardv2.ItemRecipesMap)

	// queueing it all up
	go func() {
		itemRecipes := blizzardv2.ItemRecipesMap{}
		for job := range getEncodedRecipesOut {
			if job.Err() != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve recipe")

				continue
			}

			persistRecipesIn <- ProfessionsDatabase.PersistEncodedRecipesInJob{
				RecipeId:              job.Id(),
				EncodedRecipe:         job.EncodedRecipe(),
				EncodedNormalizedName: job.EncodedNormalizedName(),
			}

			itemRecipes = itemRecipes.Merge(job.ItemRecipesMap())
		}

		close(persistRecipesIn)

		itemRecipesOut <- itemRecipes

		close(itemRecipesOut)
	}()

	// waiting for persist-recipes to finish out
	totalPersisted, err := sta.ProfessionsDatabase.PersistEncodedRecipes(persistRecipesIn)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist recipe")

		return RecipesIntakeResponse{}, err
	}

	// persisting item-recipe-ids
	itemRecipes := <-itemRecipesOut
	if err := sta.ProfessionsDatabase.PersistItemRecipes(itemRecipes); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist item-recipes")

		return RecipesIntakeResponse{}, err
	}

	// gathering recipe item-ids
	recipeItemIds, err := sta.ProfessionsDatabase.GetRecipeItemIds()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to gather recipe item-ids")

		return RecipesIntakeResponse{}, err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in recipes-intake")

	resp := RecipesIntakeResponse{
		RecipeItemIds: recipeItemIds,
	}

	if totalPersisted == 0 {
		if err := sta.ProfessionsDatabase.SetIsComplete(string(professionsflags.Recipes)); err != nil {
			return RecipesIntakeResponse{}, err
		}

		return resp, nil
	}

	return resp, nil
}
