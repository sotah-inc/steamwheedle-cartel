package items

import (
	"strings"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (idBase Database) FindMatchingItemFromRecipes(
	tuple blizzardv2.VersionItemTuple,
	rsMap blizzardv2.RecipeSubjectMap,
) (blizzardv2.RecipeIds, error) {
	out := blizzardv2.RecipeIds{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		v := bkt.Get(baseKeyName(tuple))
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

		for recipeId, recipeSubject := range rsMap {
			if !strings.Contains(foundDescription, recipeSubject) {
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
