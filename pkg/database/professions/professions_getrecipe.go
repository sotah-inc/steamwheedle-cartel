package professions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipe(id blizzardv2.RecipeId) (sotah.Recipe, error) {
	out := sotah.Recipe{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		recipesBucket := tx.Bucket(recipesBucketName())
		if recipesBucket == nil {
			return errors.New("recipes bucket was blank")
		}

		data := recipesBucket.Get(recipeKeyName(id))
		if data == nil {
			return errors.New("recipe not found")
		}

		var err error
		out, err = sotah.NewRecipe(data)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.Recipe{}, err
	}

	return out, nil
}
