package professions

import (
	"errors"
	"strings"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) FindMatchingRecipeFromItems(
	id blizzardv2.RecipeId,
	isMap blizzardv2.ItemSubjectsMap,
) (blizzardv2.ItemIds, error) {
	out := blizzardv2.ItemIds{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		recipesBucket := tx.Bucket(recipesBucketName())
		if recipesBucket == nil {
			return errors.New("recipes bucket was blank")
		}

		data := recipesBucket.Get(recipeKeyName(id))
		if data == nil {
			return nil
		}

		recipe, err := sotah.NewRecipe(data)
		if err != nil {
			return err
		}

		foundName := recipe.BlizzardMeta.Name.ResolveDefaultName()
		if foundName == "" {
			return nil
		}

		for itemId, itemSubject := range isMap {
			if !strings.Contains(itemSubject, foundName) {
				continue
			}

			out = append(out, itemId)
		}

		return nil
	})
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	return out, nil
}
