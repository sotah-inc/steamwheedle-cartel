package professions

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetRecipesJob struct {
	Err    error
	Id     blizzardv2.RecipeId
	Recipe sotah.Recipe
}

func (job GetRecipesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"recipe": job.Id,
	}
}

func (pdBase Database) GetRecipes(ids []blizzardv2.RecipeId) chan GetRecipesJob {
	in := make(chan blizzardv2.RecipeId)
	out := make(chan GetRecipesJob)

	worker := func() {
		for id := range in {
			recipe, err := pdBase.GetRecipe(id)
			if err != nil {
				out <- GetRecipesJob{
					Err:    err,
					Id:     id,
					Recipe: sotah.Recipe{},
				}

				continue
			}

			out <- GetRecipesJob{
				Err:    nil,
				Id:     id,
				Recipe: recipe,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	return out
}
