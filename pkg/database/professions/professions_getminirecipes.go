package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetMiniRecipes() (sotah.MiniRecipes, error) {
	var out sotah.MiniRecipes

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		recipesBucket := tx.Bucket(recipesBucketName())
		if recipesBucket == nil {
			return nil
		}

		return recipesBucket.ForEach(func(k []byte, v []byte) error {
			recipe, err := sotah.NewRecipe(v)
			if err != nil {
				return err
			}

			out = append(out, sotah.NewMiniRecipe(recipe))

			return nil
		})
	})
	if err != nil {
		return sotah.MiniRecipes{}, err
	}

	return out, nil
}
