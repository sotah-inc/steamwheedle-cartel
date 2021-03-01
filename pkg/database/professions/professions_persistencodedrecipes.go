package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type PersistEncodedRecipesInJob struct {
	RecipeId              blizzardv2.RecipeId
	EncodedRecipe         []byte
	EncodedNormalizedName []byte
}

func (pdBase Database) PersistEncodedRecipes(in chan PersistEncodedRecipesInJob) (int, error) {
	logging.Info("persisting encoded recipes")

	totalPersisted := 0

	err := pdBase.db.Batch(func(tx *bolt.Tx) error {
		rBucket, err := tx.CreateBucketIfNotExists(recipesBucketName())
		if err != nil {
			return err
		}

		rNameBucket, err := tx.CreateBucketIfNotExists(recipeNamesBucketName())
		if err != nil {
			return err
		}

		for job := range in {
			if err := rBucket.Put(recipeKeyName(job.RecipeId), job.EncodedRecipe); err != nil {
				return err
			}

			if err := rNameBucket.Put(
				recipeNameKeyName(job.RecipeId),
				job.EncodedNormalizedName,
			); err != nil {
				return err
			}

			totalPersisted += 1
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return totalPersisted, nil
}
