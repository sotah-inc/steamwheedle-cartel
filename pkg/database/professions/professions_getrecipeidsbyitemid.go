package professions

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/professions/itemrecipekind" // nolint:lll
)

func (pdBase Database) GetRecipeIdsByItemId(
	kind itemrecipekind.ItemRecipeKind,
	itemId blizzardv2.ItemId,
) ([]blizzardv2.RecipeId, error) {
	var out []blizzardv2.RecipeId

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(itemRecipesBucketName(kind))
		if bkt == nil {
			return nil
		}

		v := bkt.Get(itemRecipesKeyName(itemId))
		if v == nil {
			return nil
		}

		return json.Unmarshal(v, &out)
	})
	if err != nil {
		return []blizzardv2.RecipeId{}, err
	}

	return out, nil
}
