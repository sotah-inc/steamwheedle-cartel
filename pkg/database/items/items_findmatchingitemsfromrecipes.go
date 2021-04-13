package items

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type FindMatchingItemsFromRecipesJob struct {
	Err       error
	Id        blizzardv2.ItemId
	RecipeIds []blizzardv2.RecipeId
}

func (idBase Database) FindMatchingItemsFromRecipes(
	rsMap blizzardv2.RecipeSubjectMap,
) (blizzardv2.ItemRecipesMap, error) {
	// resolving all item-ids
	ids, err := idBase.GetItemIds()
	if err != nil {
		return blizzardv2.ItemRecipesMap{}, err
	}

	// establish channels
	in := make(chan blizzardv2.ItemId)
	out := make(chan FindMatchingItemsFromRecipesJob)

	// spinning up workers
	worker := func() {
		for id := range in {
			recipeIds, err := idBase.FindMatchingItemFromRecipes(id, rsMap)
			if err != nil {
				out <- FindMatchingItemsFromRecipesJob{
					Err:       err,
					Id:        id,
					RecipeIds: []blizzardv2.RecipeId{},
				}

				continue
			}

			out <- FindMatchingItemsFromRecipesJob{
				Err:       err,
				Id:        id,
				RecipeIds: recipeIds,
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

	results := blizzardv2.ItemRecipesMap{}
	for job := range out {
		if job.Err != nil {
			return blizzardv2.ItemRecipesMap{}, job.Err
		}

		if len(job.RecipeIds) == 0 {
			continue
		}

		results[job.Id] = job.RecipeIds
	}

	return results, nil
}
