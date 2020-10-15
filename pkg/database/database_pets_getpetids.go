package database

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (pdBase PetsDatabase) GetPetIds() ([]blizzardv2.PetId, error) {
	var out []blizzardv2.PetId

	// peeking into the items database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		petsBucket := tx.Bucket(databasePetsBucketName())
		if petsBucket == nil {
			return nil
		}

		return petsBucket.ForEach(func(k []byte, v []byte) error {
			petId, err := petIdFromPetKeyName(k)
			if err != nil {
				return err
			}

			out = append(out, petId)

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.PetId{}, err
	}

	return out, nil
}
