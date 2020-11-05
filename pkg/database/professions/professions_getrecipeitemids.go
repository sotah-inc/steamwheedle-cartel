package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipeItemIds() (blizzardv2.ItemIds, error) {
	results := blizzardv2.ItemIdsMap{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		recipesBucket := tx.Bucket(recipesBucketName())
		if recipesBucket == nil {
			return nil
		}

		return recipesBucket.ForEach(func(key []byte, value []byte) error {
			recipe, err := sotah.NewRecipe(value)
			if err != nil {
				return err
			}

			for _, id := range recipe.ItemIds() {
				results[id] = struct{}{}
			}

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.ItemId{}, err
	}

	return results.ToItemIds(), nil
}
