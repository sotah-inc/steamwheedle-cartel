package items

import (
	"strings"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (idBase Database) FindMatchingItemFromRecipes(
	id blizzardv2.ItemId,
	recipeDescriptions blizzardv2.RecipeIdDescriptionMap,
) (blizzardv2.RecipeIds, error) {
	out := blizzardv2.RecipeIds{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		v := bkt.Get(baseKeyName(id))
		if v == nil {
			return nil
		}

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

		for recipeId, recipeDescription := range recipeDescriptions {
			if !strings.Contains(foundDescription, recipeDescription) {
				continue
			}

			out = append(out, recipeId)
		}

		return nil
	})
	if err != nil {
		return blizzardv2.RecipeIds{}, err
	}

	return out, nil
}
