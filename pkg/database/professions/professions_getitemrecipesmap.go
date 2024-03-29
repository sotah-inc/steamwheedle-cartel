package professions

import (
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions/itemrecipekind" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type GetItemRecipesMapOutJob struct {
	Err       error
	ItemId    blizzardv2.ItemId
	RecipeIds []blizzardv2.RecipeId
}

func (job GetItemRecipesMapOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"item":  job.ItemId,
	}
}

func (pdBase Database) GetItemRecipesMap(
	kind itemrecipekind.ItemRecipeKind,
	ids blizzardv2.ItemIds,
) (blizzardv2.ItemRecipesMap, error) {
	// establishing channels
	in := make(chan blizzardv2.ItemId)
	out := make(chan GetItemRecipesMapOutJob)

	// spinning up workers for receiving encoded-data and persisting it
	worker := func() {
		for id := range in {
			recipeIds, err := pdBase.GetRecipeIdsByItemId(kind, id)
			if err != nil {
				out <- GetItemRecipesMapOutJob{
					Err:       err,
					ItemId:    id,
					RecipeIds: []blizzardv2.RecipeId{},
				}

				continue
			}

			out <- GetItemRecipesMapOutJob{
				Err:       nil,
				ItemId:    id,
				RecipeIds: recipeIds,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	results := blizzardv2.ItemRecipesMap{}
	for job := range out {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("failed to resolve recipe-ids")

			return blizzardv2.ItemRecipesMap{}, job.Err
		}

		results[job.ItemId] = job.RecipeIds
	}

	return results, nil
}
