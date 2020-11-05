package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (pdBase Database) GetRecipeIds() ([]blizzardv2.RecipeId, error) {
	var out []blizzardv2.RecipeId

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		recipesBucket := tx.Bucket(recipesBucketName())
		if recipesBucket == nil {
			return nil
		}

		return recipesBucket.ForEach(func(key []byte, value []byte) error {
			parsedId, err := recipeIdFromKeyName(key)
			if err != nil {
				return err
			}

			out = append(out, parsedId)

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.RecipeId{}, err
	}

	return out, nil
}
