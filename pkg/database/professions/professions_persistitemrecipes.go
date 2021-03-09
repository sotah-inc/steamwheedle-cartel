package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (pdBase Database) PersistItemRecipes(providedItemRecipes blizzardv2.ItemRecipesMap) error {
	currentItemRecipes, err := pdBase.GetItemRecipesMap(providedItemRecipes.ItemIds())
	if err != nil {
		return err
	}

	nextItemRecipes := currentItemRecipes.Merge(providedItemRecipes)

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
