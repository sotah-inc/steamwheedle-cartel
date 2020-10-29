package pets

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type PersistEncodedPetsInJob struct {
	Id                    blizzardv2.PetId
	EncodedPet            []byte
	EncodedNormalizedName []byte
}

func (pdBase Database) PersistEncodedPets(
	in chan PersistEncodedPetsInJob,
) (int, error) {
	logging.Info("persisting encoded pets")

	totalPersisted := 0

	err := pdBase.db.Batch(func(tx *bolt.Tx) error {
		petsBucket, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		petNamesBucket, err := tx.CreateBucketIfNotExists(namesBucketName())
		if err != nil {
			return err
		}

		for job := range in {
			if err := petsBucket.Put(baseKeyName(job.Id), job.EncodedPet); err != nil {
				return err
			}

			if err := petNamesBucket.Put(nameKeyName(job.Id), job.EncodedNormalizedName); err != nil {
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
