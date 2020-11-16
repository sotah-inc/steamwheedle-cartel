package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipes(idList []blizzardv2.RecipeId) ([]sotah.Recipe, error) {
	var out []sotah.Recipe

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

			out = append(out, recipe)
		}

		return nil
	})
	if err != nil {
		return []sotah.Recipe{}, err
	}

	return out, nil
}
