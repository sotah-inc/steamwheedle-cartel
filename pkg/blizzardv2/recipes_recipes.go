package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetRecipesOptions struct {
	GetRecipeURL func(RecipeId) (string, error)

	RecipesGroup RecipesGroup
	Limit        int
}

type GetRecipesOutJob struct {
	Err            error
	ProfessionId   ProfessionId
	SkillTierId    SkillTierId
	Id             RecipeId
	RecipeResponse RecipeResponse
}

type GetRecipesInJob struct {
	ProfessionId ProfessionId
	SkillTierId  SkillTierId
	Id           RecipeId
}

func (job GetRecipesOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"id":    job.Id,
	}
}

func GetRecipes(opts GetRecipesOptions) chan GetRecipesOutJob {
	// starting up workers for gathering individual recipes
	in := make(chan GetRecipesInJob)
	out := make(chan GetRecipesOutJob)
	worker := func() {
		for inJob := range in {
			getRecipeUri, err := opts.GetRecipeURL(inJob.Id)
			if err != nil {
				out <- GetRecipesOutJob{
					Err:            err,
					ProfessionId:   inJob.ProfessionId,
					SkillTierId:    inJob.SkillTierId,
					Id:             inJob.Id,
					RecipeResponse: RecipeResponse{},
				}

				continue
			}

			recipeResp, _, err := NewRecipeResponseFromHTTP(getRecipeUri)
			if err != nil {
				out <- GetRecipesOutJob{
					Err:            err,
					ProfessionId:   inJob.ProfessionId,
					SkillTierId:    inJob.SkillTierId,
					Id:             inJob.Id,
					RecipeResponse: RecipeResponse{},
				}

				continue
			}

			out <- GetRecipesOutJob{
				Err:            nil,
				ProfessionId:   inJob.ProfessionId,
				SkillTierId:    inJob.SkillTierId,
				Id:             inJob.Id,
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
		for professionId, skillTiersGroup := range opts.RecipesGroup {
			for skillTierId, recipeIds := range skillTiersGroup {
				for _, id := range recipeIds {
					in <- GetRecipesInJob{
						ProfessionId: professionId,
						SkillTierId:  skillTierId,
						Id:           id,
					}

					total += 1

					if total > opts.Limit {
						close(in)

						return
					}
				}
			}
		}

		close(in)
	}()

	return out
}
