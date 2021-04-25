package professions

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type FindMatchingRecipesFromItemsJob struct {
	Err     error
	Id      blizzardv2.RecipeId
	ItemIds blizzardv2.ItemIds
}

func (pdBase Database) FindMatchingRecipesFromItems(
	isMap blizzardv2.ItemSubjectsMap,
) (blizzardv2.ItemRecipesMap, error) {
	// resolving all recipe-ids
	ids, err := pdBase.GetRecipeIds()
	if err != nil {
		return blizzardv2.ItemRecipesMap{}, err
	}

	// establish channels
	in := make(chan blizzardv2.RecipeId)
	out := make(chan FindMatchingRecipesFromItemsJob)

	// spinning up workers
	worker := func() {
		for id := range in {
			itemIds, err := pdBase.FindMatchingRecipeFromItems(id, isMap)
			if err != nil {
				out <- FindMatchingRecipesFromItemsJob{
					Err:     err,
					Id:      id,
					ItemIds: blizzardv2.ItemIds{},
				}

				continue
			}

			out <- FindMatchingRecipesFromItemsJob{
				Err:     err,
				Id:      id,
				ItemIds: itemIds,
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

		if len(job.ItemIds) == 0 {
			continue
		}

		for _, itemId := range job.ItemIds {
			results[itemId] = results.Find(itemId).Merge(blizzardv2.RecipeIds{job.Id})
		}
	}

	return results, nil
}
