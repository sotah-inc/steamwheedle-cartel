package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipeDescriptions(
	idList []blizzardv2.RecipeId,
) (blizzardv2.RecipeIdDescriptionMap, error) {
	out := blizzardv2.RecipeIdDescriptionMap{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(recipesBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range idList {
			v := bkt.Get(recipeKeyName(id))
			if v == nil {
				continue
			}

			recipe, err := sotah.NewRecipe(v)
			if err != nil {
				return err
			}

			defaultDescription := recipe.BlizzardMeta.Description.ResolveDefaultName()
			if defaultDescription == "" {
				return nil
			}

			out[id] = defaultDescription
		}

		return nil
	})
	if err != nil {
		return blizzardv2.RecipeIdDescriptionMap{}, err
	}

	return out, nil
}
