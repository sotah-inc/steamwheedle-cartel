package items

import (
	"strings"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (idBase Database) FindMatchingFromRecipes(
	recipeDescriptions blizzardv2.RecipeIdDescriptionMap,
) (blizzardv2.ItemRecipesMap, error) {
	out := blizzardv2.ItemRecipesMap{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			item, err := sotah.NewItemFromGzipped(v)
			if err != nil {
				return err
			}

			foundDescription := func() string {
				if len(item.BlizzardMeta.PreviewItem.Spells) == 0 {
					return ""
				}

				return item.BlizzardMeta.PreviewItem.Spells[0].Description.ResolveDefaultName()
			}()
			if foundDescription == "" {
				return nil
			}

			out[item.BlizzardMeta.Id] = blizzardv2.RecipeIds{}
			for recipeId, recipeDescription := range recipeDescriptions {
				if !strings.Contains(foundDescription, recipeDescription) {
					continue
				}

				out[item.BlizzardMeta.Id] = append(out[item.BlizzardMeta.Id], recipeId)
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
