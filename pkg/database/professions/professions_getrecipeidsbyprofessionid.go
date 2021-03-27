package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipeIdsByProfessionId(
	professionId blizzardv2.ProfessionId,
) ([]blizzardv2.RecipeId, error) {
	var out []blizzardv2.RecipeId

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		stBucket := tx.Bucket(skillTiersBucketName(professionId))
		if stBucket == nil {
			return nil
		}

		return stBucket.ForEach(func(k []byte, v []byte) error {
			skillTier, err := sotah.NewSkillTier(v)
			if err != nil {
				return err
			}

			for _, category := range skillTier.BlizzardMeta.Categories {
				for _, categoryRecipe := range category.Recipes {
					out = append(out, categoryRecipe.Id)
				}
			}

			return nil
		})

	})
	if err != nil {
		return []blizzardv2.RecipeId{}, err
	}

	return out, nil
}
