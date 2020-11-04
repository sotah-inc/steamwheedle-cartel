package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetRecipesOptions struct {
	GetRecipeURL func(RecipeId) (string, error)

	RecipeIds []RecipeId
	Limit     int
}

type GetRecipesJob struct {
	Err            error
	Id             RecipeId
	RecipeResponse RecipeResponse
}

func (job GetRecipesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func GetRecipes(opts GetRecipesOptions) chan GetRecipesJob {
	// starting up workers for gathering individual recipes
	in := make(chan RecipeId)
	out := make(chan GetRecipesJob)
	worker := func() {
		for id := range in {
			getRecipeUri, err := opts.GetRecipeURL(id)
			if err != nil {
				out <- GetRecipesJob{
					Err:            err,
					Id:             id,
					RecipeResponse: RecipeResponse{},
				}

				continue
			}

			recipeResp, _, err := NewRecipeResponseFromHTTP(getRecipeUri)
			if err != nil {
				out <- GetRecipesJob{
					Err:            err,
					Id:             id,
					RecipeResponse: RecipeResponse{},
				}

				continue
			}

			out <- GetRecipesJob{
				Err:            nil,
				Id:             id,
				RecipeResponse: recipeResp,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		total := 0
		for _, id := range opts.RecipeIds {
			in <- id

			total += 1

			if total > opts.Limit {
				break
			}
		}

		close(in)
	}()

	return out
}
