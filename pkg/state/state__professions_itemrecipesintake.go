package state

import (
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	ProfessionsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
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

	persistEncodedRecipesIn := make(chan ProfessionsDatabase.PersistEncodedRecipesInJob)
	go func() {
		recipeItems := irMap.ToRecipesItemMap()
		for getRecipesOutJob := range sta.ProfessionsDatabase.GetRecipes(irMap.RecipeIds()) {
			if getRecipesOutJob.Err != nil {
				logging.WithFields(getRecipesOutJob.ToLogrusFields()).Error("failed to fetch recipe")

				continue
			}

			supplementalItemId, ok := recipeItems[getRecipesOutJob.Id]
			if !ok {
				logging.WithField("recipe", getRecipesOutJob.Id).Info("no item for recipe")

				continue
			}

			logging.WithFields(logrus.Fields{
				"recipe": getRecipesOutJob.Id,
				"item":   supplementalItemId,
			}).Info("updating crafted-item for recipe")

			nextRecipe := sotah.Recipe{
				BlizzardMeta: blizzardv2.RecipeResponse{
					LinksBase:   getRecipesOutJob.Recipe.BlizzardMeta.LinksBase,
					Id:          getRecipesOutJob.Recipe.BlizzardMeta.Id,
					Name:        getRecipesOutJob.Recipe.BlizzardMeta.Name,
					Description: getRecipesOutJob.Recipe.BlizzardMeta.Description,
					Media:       getRecipesOutJob.Recipe.BlizzardMeta.Media,
					CraftedItem: blizzardv2.RecipeItem{
						Key:  blizzardv2.HrefReference{},
						Name: map[locale.Locale]string{},
						Id:   supplementalItemId,
					},
					AllianceCraftedItem:   getRecipesOutJob.Recipe.BlizzardMeta.AllianceCraftedItem,
					HordeCraftedItem:      getRecipesOutJob.Recipe.BlizzardMeta.HordeCraftedItem,
					Reagents:              getRecipesOutJob.Recipe.BlizzardMeta.Reagents,
					Rank:                  getRecipesOutJob.Recipe.BlizzardMeta.Rank,
					CraftedQuantity:       getRecipesOutJob.Recipe.BlizzardMeta.CraftedQuantity,
					ModifiedCraftingSlots: getRecipesOutJob.Recipe.BlizzardMeta.ModifiedCraftingSlots,
				},
				SotahMeta: getRecipesOutJob.Recipe.SotahMeta,
			}

			encodedNextRecipe, err := nextRecipe.EncodeForStorage()
			if err != nil {
				logging.WithField("error", err.Error()).Error("failed to encode next-recipe for storage")

				continue
			}

			persistEncodedRecipesIn <- ProfessionsDatabase.PersistEncodedRecipesInJob{
				RecipeId:              getRecipesOutJob.Id,
				EncodedRecipe:         encodedNextRecipe,
				EncodedNormalizedName: nil,
			}
		}

		close(persistEncodedRecipesIn)
	}()

	totalPersisted, err := sta.ProfessionsDatabase.PersistEncodedRecipes(persistEncodedRecipesIn)
	if err != nil {
		return err
	}

	logging.WithFields(logrus.Fields{
		"total":          totalPersisted,
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("persisted recipes")

	return nil
}
