package disk

import (
	"errors"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type getEncodedRecipeJob struct {
	err                   error
	id                    blizzardv2.RecipeId
	itemRecipesMap        blizzardv2.ItemRecipesMap
	encodedRecipe         []byte
	encodedNormalizedName []byte
}

func (g getEncodedRecipeJob) Err() error                                { return g.err }
func (g getEncodedRecipeJob) Id() blizzardv2.RecipeId                   { return g.id }
func (g getEncodedRecipeJob) ItemRecipesMap() blizzardv2.ItemRecipesMap { return g.itemRecipesMap }
func (g getEncodedRecipeJob) EncodedRecipe() []byte                     { return g.encodedRecipe }
func (g getEncodedRecipeJob) EncodedNormalizedName() []byte {
	return g.encodedNormalizedName
}
func (g getEncodedRecipeJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": g.err.Error(),
		"id":    g.id,
	}
}

func (client Client) GetEncodedRecipes(
	recipesGroup blizzardv2.RecipesGroup,
) chan BaseLake.GetEncodedRecipeJob {
	out := make(chan BaseLake.GetEncodedRecipeJob)

	// starting up workers for resolving recipes
	recipesOut := client.resolveRecipes(recipesGroup)

	// starting up workers for resolving recipe-medias
	recipeMediasIn := make(chan blizzardv2.GetRecipeMediasInJob)
	recipeMediasOut := client.resolveRecipeMedias(recipeMediasIn)

	// queueing it all up
	go func() {
		for job := range recipesOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve recipe")

				continue
			}

			logging.WithField(
				"recipe-id", job.RecipeResponse.Id,
			).Info("enqueueing recipe for recipe-media resolution")

			recipeMediasIn <- blizzardv2.GetRecipeMediasInJob{
				RecipeResponse: job.RecipeResponse,
				ProfessionId:   job.ProfessionId,
				SkillTierId:    job.SkillTierId,
			}
		}

		close(recipeMediasIn)
	}()
	go func() {
		for job := range recipeMediasOut {
			if job.Err != nil {
				logging.WithFields(job.ToLogrusFields()).Error("failed to resolve recipe")

				continue
			}

			recipeIconUrl, err := job.RecipeMediaResponse.GetIconUrl()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":    err.Error(),
					"response": job.RecipeMediaResponse,
				}).Error("recipe-media did not have icon")

				continue
			}

			recipe := sotah.Recipe{
				BlizzardMeta: job.RecipeResponse,
				SotahMeta: sotah.RecipeMeta{
					ProfessionId: job.ProfessionId,
					SkillTierId:  job.SkillTierId,
					IconUrl:      recipeIconUrl,
				},
			}

			encodedRecipe, err := recipe.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"recipe": recipe.BlizzardMeta.Id,
				}).Error("failed to encode recipe for storage")

				continue
			}

			normalizedName, err := func() (locale.Mapping, error) {
				foundName, ok := job.RecipeResponse.Name[locale.EnUS]
				if !ok {
					return locale.Mapping{}, errors.New("failed to resolve enUS name")
				}

				normalizedName, err := sotah.NormalizeString(foundName)
				if err != nil {
					return locale.Mapping{}, err
				}

				return locale.Mapping{locale.EnUS: normalizedName}, nil
			}()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"recipe": recipe.BlizzardMeta.Id,
				}).Error("failed to normalize name")

				continue
			}

			encodedNormalizedName, err := normalizedName.EncodeForStorage()
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"recipe": recipe.BlizzardMeta.Id,
				}).Error("failed to encode normalized-name for storage")

				continue
			}

			craftedItemIds := blizzardv2.ItemIds{
				job.RecipeResponse.CraftedItem.Id,
				job.RecipeResponse.HordeCraftedItem.Id,
				job.RecipeResponse.AllianceCraftedItem.Id,
			}
			itemRecipesMap := blizzardv2.ItemRecipesMap{}
			for _, id := range craftedItemIds {
				itemRecipesMap = itemRecipesMap.Merge(blizzardv2.ItemRecipesMap{
					id: blizzardv2.RecipeIds{job.RecipeResponse.Id},
				})
			}
			for _, id := range job.RecipeResponse.ReagentItemIds() {
				itemRecipesMap = itemRecipesMap.Merge(blizzardv2.ItemRecipesMap{
					id: blizzardv2.RecipeIds{job.RecipeResponse.Id},
				})
			}

			out <- getEncodedRecipeJob{
				err:                   nil,
				id:                    job.RecipeResponse.Id,
				itemRecipesMap:        itemRecipesMap,
				encodedRecipe:         encodedRecipe,
				encodedNormalizedName: encodedNormalizedName,
			}
		}

		close(out)
	}()

	return out
}
