package professions

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (pdBase Database) PersistItemRecipes(providedItemRecipes blizzardv2.ItemRecipesMap) error {
	if len(providedItemRecipes) == 0 {
		logging.Info("skipping persisting item-recipes due to no provided item-recipes")

		return nil
	}

	currentItemRecipes, err := pdBase.GetItemRecipesMap(providedItemRecipes.ItemIds())
	if err != nil {
		return err
	}

	nextItemRecipes := currentItemRecipes.Merge(providedItemRecipes)

	logging.WithFields(logrus.Fields{
		"provided-items": providedItemRecipes,
		"current-items":  currentItemRecipes,
		"next-items":     nextItemRecipes,
	}).Info("persisting item-recipes")

	return pdBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(itemsCraftedByBucketName())
		if err != nil {
			return err
		}

		for itemId, recipeIds := range nextItemRecipes {
			encodedRecipeIds, err := recipeIds.EncodeForStorage()
			if err != nil {
				return err
			}

			if err := bkt.Put(itemsCraftedByKeyName(itemId), encodedRecipeIds); err != nil {
				return err
			}
		}

		return nil
	})
}
