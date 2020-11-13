package disk

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

type getEncodedRecipeJob struct {
	err           error
	id            blizzardv2.RecipeId
	encodedRecipe []byte
}

func (g getEncodedRecipeJob) Err() error              { return g.err }
func (g getEncodedRecipeJob) Id() blizzardv2.RecipeId { return g.id }
func (g getEncodedRecipeJob) EncodedRecipe() []byte   { return g.encodedRecipe }
func (g getEncodedRecipeJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": g.err.Error(),
		"id":    g.id,
	}
}

func (client Client) GetEncodedRecipes(
	ids []blizzardv2.RecipeId,
) chan BaseLake.GetEncodedRecipeJob {
	out := make(chan BaseLake.GetEncodedRecipeJob)

	// starting up workers for resolving recipes
	recipesOut := client.resolveRecipes(ids)

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
			).Info("enqueueing profession for recipe-media resolution")

			recipeMediasIn <- blizzardv2.GetRecipeMediasInJob{
				RecipeResponse: job.RecipeResponse,
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
					IconUrl: recipeIconUrl,
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

			out <- getEncodedRecipeJob{
				err:           nil,
				id:            job.RecipeResponse.Id,
				encodedRecipe: encodedRecipe,
			}
		}

		close(out)
	}()

	return out
}
