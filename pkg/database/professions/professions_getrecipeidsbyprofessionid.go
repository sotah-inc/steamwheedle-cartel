package professions

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetRecipeIdsByProfessionId(
	professionId blizzardv2.ProfessionId,
) ([]blizzardv2.RecipeId, error) {
	var out []blizzardv2.RecipeId

	logging.WithFields(logrus.Fields{
		"profession": professionId,
	}).Info("checking for recipe-ids by profession")

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		stBucket := tx.Bucket(skillTiersBucketName(professionId))
		if stBucket == nil {
			logging.WithFields(logrus.Fields{
				"profession": professionId,
			}).Info("skill-tiers bucket was empty")

			return nil
		}

		return stBucket.ForEach(func(k []byte, v []byte) error {
			skillTier, err := sotah.NewSkillTier(v)
			if err != nil {
				return err
			}

			logging.WithFields(logrus.Fields{
				"profession": professionId,
				"skill-tier": skillTier,
			}).Info("checking skill-tier")

			for _, category := range skillTier.BlizzardMeta.Categories {
				logging.WithFields(logrus.Fields{
					"profession": professionId,
					"skill-tier": skillTier,
					"category":   category,
				}).Info("checking category")

				for _, categoryRecipe := range category.Recipes {
					logging.WithFields(logrus.Fields{
						"profession": professionId,
						"skill-tier": skillTier,
						"category":   category,
						"recipe":     categoryRecipe.Id,
					}).Info("appending category recipe")

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
