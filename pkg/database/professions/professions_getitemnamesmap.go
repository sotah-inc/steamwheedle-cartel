package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetIdNormalizedNameMap() (sotah.RecipeIdNameMap, error) {
	out := sotah.RecipeIdNameMap{}

	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(recipeNamesBucketName())
		if bkt == nil {
			return nil
		}

		return bkt.ForEach(func(k, v []byte) error {
			recipeId, err := recipeIdFromNameKeyName(k)
			if err != nil {
				logging.WithField(
					"error",
					err.Error(),
				).Error("failed to call recipeIdFromNameKeyName")

				return err
			}

			out[recipeId], err = locale.NewMapping(v)
			if err != nil {
				logging.WithField(
					"error",
					err.Error(),
				).Error("failed to call locale.NewMapping")

				return err
			}

			return nil
		})
	})
	if err != nil {
		return sotah.RecipeIdNameMap{}, err
	}

	return out, nil
}
