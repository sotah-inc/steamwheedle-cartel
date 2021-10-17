package pets

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewDatabase(dbDir string) (Database, error) {
	dbFilepath, err := DatabasePath(dbDir)
	if err != nil {
		return Database{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("initializing pets database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return Database{}, err
	}

	return Database{db}, nil
}

type Database struct {
	db *bolt.DB
}

// gathering pets
func (pdBase Database) GetPets() ([]sotah.Pet, error) {
	var out []sotah.Pet

	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
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

func (pdBase Database) GetIdNormalizedNameMap() (sotah.PetIdNameMap, error) {
	out := sotah.PetIdNameMap{}

	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(namesBucketName())
		if bkt == nil {
			return nil
		}

		err := bkt.ForEach(func(k, v []byte) error {
			petId, err := petIdFromNameKeyName(k)
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

func (pdBase Database) FindPets(petIds []blizzardv2.PetId) ([]sotah.Pet, error) {
	var out []sotah.Pet
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			return nil
		}

		for _, id := range petIds {
			value := bkt.Get(baseKeyName(id))
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

func (pdBase Database) IsComplete() (bool, error) {
	out := false

	err := pdBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(flagsBucketName())
		if bkt == nil {
			return nil
		}

		data := bkt.Get(isCompleteKeyName())
		if data == nil {
			return nil
		}

		if string(data) == "1" {
			out = true
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return out, nil
}

func (pdBase Database) SetIsComplete() error {
	return pdBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(flagsBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(isCompleteKeyName(), []byte("1")); err != nil {
			return err
		}

		return nil
	})
}
