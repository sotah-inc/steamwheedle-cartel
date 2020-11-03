package professions

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/boltdb/bolt"
)

func (pdBase Database) GetProfessions() ([]sotah.Profession, error) {
	var out []sotah.Profession

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return nil
		}

		return baseBucket.ForEach(func(key []byte, value []byte) error {
			data := baseBucket.Get(key)
			if data == nil {
				return errors.New("profession not found")
			}

			foundProfession, err := sotah.NewProfession(data)
			if err != nil {
				return err
			}

			out = append(out, foundProfession)

			return nil
		})

	})
	if err != nil {
		return []sotah.Profession{}, err
	}

	return out, nil
}
