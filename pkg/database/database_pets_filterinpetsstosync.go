package database

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (pdBase PetsDatabase) FilterInPetsToSync(ids []blizzardv2.PetId) ([]blizzardv2.PetId, error) {
	var out []blizzardv2.PetId

	// peeking into the items database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		petsBucket, err := tx.CreateBucketIfNotExists(databasePetsBucketName())
		if err != nil {
			return err
		}

		for _, id := range ids {
			value := petsBucket.Get(petsKeyName(id))
			if value != nil {
				continue
			}

			out = append(out, id)
		}

		return nil
	})
	if err != nil {
		return []blizzardv2.PetId{}, err
	}

	return out, nil
}
