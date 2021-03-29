package items

import (
	"strings"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func (idBase Database) FindMatchingFromRecipeNames(
	recipeNormalizedNames blizzardv2.RecipeIdNameMap,
) (blizzardv2.ItemRecipesMap, error) {
	out := blizzardv2.ItemRecipesMap{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(namesBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			itemId, err := itemIdFromNameKeyName(k)
			if err != nil {
				return err
			}

			mapping, err := locale.NewMapping(v)
			if err != nil {
				return err
			}

			itemNormalizedName := mapping.ResolveDefaultName()
			if itemNormalizedName == "" {
				return nil
			}

			out[itemId] = blizzardv2.RecipeIds{}
			for recipeId, recipeNormalizedName := range recipeNormalizedNames {
				if !strings.Contains(itemNormalizedName, recipeNormalizedName) {
					continue
				}

				out[itemId] = append(out[itemId], recipeId)
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return blizzardv2.ItemRecipesMap{}, err
	}

	return out, nil
}
