package pricelisthistory

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (phdBase Database) getRecipePrices(id blizzardv2.RecipeId) (sotah.RecipePrices, error) {
	out := sotah.RecipePrices{}

	err := phdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(recipeBucketName(id))
		if bkt == nil {
			return nil
		}

		value := bkt.Get(recipeKeyName())
		if value == nil {
			return nil
		}

		var err error
		out, err = sotah.NewRecipePricesFromGzip(value)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.RecipePrices{}, err
	}

	return out, nil

}
