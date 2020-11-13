package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipesIcons(idList []blizzardv2.RecipeId) (map[blizzardv2.RecipeId]string, error) {
	out := map[blizzardv2.RecipeId]string{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		recipesBucket := tx.Bucket(recipesBucketName())
		if recipesBucket == nil {
			return nil
		}

		for _, id := range idList {
			value := recipesBucket.Get(recipeKeyName(id))
			if value == nil {
				continue
			}

			recipe, err := sotah.NewRecipe(value)
			if err != nil {
				return err
			}

			out[recipe.BlizzardMeta.Id] = recipe.SotahMeta.IconUrl
		}

		return nil
	})
	if err != nil {
		return map[blizzardv2.RecipeId]string{}, err
	}

	return out, nil
}
