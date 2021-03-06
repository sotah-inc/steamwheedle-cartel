package blizzardv2

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetRecipeMediasInJob struct {
	RecipeResponse RecipeResponse
	ProfessionId   ProfessionId
	SkillTierId    SkillTierId
}

type GetRecipeMediasOutJob struct {
	Err                 error
	RecipeResponse      RecipeResponse
	RecipeMediaResponse RecipeMediaResponse
	ProfessionId        ProfessionId
	SkillTierId         SkillTierId
}

func (job GetRecipeMediasOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error":  job.Err.Error(),
		"recipe": job.RecipeResponse.Id,
	}
}

func GetRecipeMedias(
	in chan GetRecipeMediasInJob,
	getRecipeMediaURL func(string) (string, error),
) chan GetRecipeMediasOutJob {
	// starting up workers for gathering individual recipes
	out := make(chan GetRecipeMediasOutJob)
	worker := func() {
		for job := range in {
			recipeMediaUrl, err := getRecipeMediaURL(job.RecipeResponse.Media.Key.Href)
			if err != nil {
				out <- GetRecipeMediasOutJob{
					Err:                 err,
					RecipeResponse:      job.RecipeResponse,
					RecipeMediaResponse: RecipeMediaResponse{},
					ProfessionId:        job.ProfessionId,
					SkillTierId:         job.SkillTierId,
				}

				continue
			}

			recipeMediaResponse, _, err := NewRecipeMediaResponseFromHTTP(recipeMediaUrl)
			if err != nil {
				out <- GetRecipeMediasOutJob{
					Err:                 err,
					RecipeResponse:      job.RecipeResponse,
					RecipeMediaResponse: RecipeMediaResponse{},
					ProfessionId:        job.ProfessionId,
					SkillTierId:         job.SkillTierId,
				}

				continue
			}

			out <- GetRecipeMediasOutJob{
				Err:                 nil,
				RecipeResponse:      job.RecipeResponse,
				RecipeMediaResponse: recipeMediaResponse,
				ProfessionId:        job.ProfessionId,
				SkillTierId:         job.SkillTierId,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	return out
}
