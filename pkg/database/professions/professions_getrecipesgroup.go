package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipesGroup() (blizzardv2.RecipesGroup, error) {
	out := blizzardv2.RecipesGroup{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return nil
		}

		return baseBucket.ForEach(func(professionKey []byte, professionValue []byte) error {
			professionId, err := professionIdFromKeyName(professionKey)
			if err != nil {
				return err
			}

			out[professionId] = map[blizzardv2.SkillTierId]blizzardv2.RecipeIds{}

			skillTiersBucket := tx.Bucket(skillTiersBucketName(professionId))
			if skillTiersBucket == nil {
				return nil
			}

			return skillTiersBucket.ForEach(func(skillTiersKey []byte, skillTiersValue []byte) error {
				skillTier, err := sotah.NewSkillTier(skillTiersValue)
				if err != nil {
					return err
				}

				out[professionId][skillTier.BlizzardMeta.Id] = skillTier.RecipeIds()

				return nil
			})
		})
	})
	if err != nil {
		return blizzardv2.RecipesGroup{}, err
	}

	return out, nil
}
