package pricelisthistory

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (phdBase Database) persistEncodedRecipePrices(data map[blizzardv2.RecipeId][]byte) error {
	logging.WithField("recipes", len(data)).Info("persisting encoded recipe-prices")

	err := phdBase.db.Batch(func(tx *bolt.Tx) error {
		for recipeId, payload := range data {
			bkt, err := tx.CreateBucketIfNotExists(recipeBucketName(recipeId))
			if err != nil {
				return err
			}

			if err := bkt.Put(recipeKeyName(), payload); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
