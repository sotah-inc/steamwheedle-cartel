package database

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewPetsDatabase(dbDir string) (PetsDatabase, error) {
	dbFilepath, err := PetsDatabasePath(dbDir)
	if err != nil {
		return PetsDatabase{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing pets database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return PetsDatabase{}, err
	}

	return PetsDatabase{db}, nil
}

type PetsDatabase struct {
	db *bolt.DB
}

// gathering pets
func (pdBase PetsDatabase) GetPets() ([]sotah.Pet, error) {
	var out []sotah.Pet

	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databasePetsBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			pet, err := sotah.NewPetFromGzipped(v)
			if err != nil {
				return err
			}

			out = append(out, pet)

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return []sotah.Pet{}, err
	}

	return out, nil
}

func (pdBase PetsDatabase) GetIdNormalizedNameMap() (sotah.PetIdNameMap, error) {
	out := sotah.PetIdNameMap{}

	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databasePetNamesBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			petId, err := petIdFromPetNameKeyName(k)
			if err != nil {
				return err
			}

			out[petId], err = locale.NewMapping(v)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.PetIdNameMap{}, err
	}

	return out, nil
}

func (pdBase PetsDatabase) FindPets(petIds []blizzardv2.PetId) ([]sotah.Pet, error) {
	var out []sotah.Pet
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databasePetsBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range petIds {
			value := bkt.Get(petsKeyName(id))
			if value == nil {
				continue
			}

			pet, err := sotah.NewPetFromGzipped(value)
			if err != nil {
				return err
			}

			out = append(out, pet)
		}

		return nil
	})
	if err != nil {
		return []sotah.Pet{}, err
	}

	return out, nil
}
