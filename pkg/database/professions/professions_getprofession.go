package professions

import (
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (pdBase Database) GetProfession(id blizzardv2.ProfessionId) (sotah.Profession, error) {
	out := sotah.Profession{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return nil
		}

		data := baseBucket.Get(baseKeyName(id))
		if data == nil {
			return errors.New("profession not found")
		}

		var err error
		out, err = sotah.NewProfession(data)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.Profession{}, err
	}

	return out, nil
}
