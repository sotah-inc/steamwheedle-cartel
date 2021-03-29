package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func (pdBase Database) GetRecipeName(
	idList []blizzardv2.RecipeId,
) (blizzardv2.RecipeIdNameMap, error) {
	out := blizzardv2.RecipeIdNameMap{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(recipeNamesBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range idList {
			v := bkt.Get(recipeNameKeyName(id))
			if v == nil {
				continue
			}

			mapping, err := locale.NewMapping(v)
			if err != nil {
				return err
			}

			defaultName := mapping.ResolveDefaultName()
			if defaultName == "" {
				return nil
			}

			out[id] = defaultName
		}

		return nil
	})
	if err != nil {
		return blizzardv2.RecipeIdNameMap{}, err
	}

	return out, nil
}
