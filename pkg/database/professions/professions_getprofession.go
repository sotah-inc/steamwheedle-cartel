package professions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetProfession(id blizzardv2.ProfessionId) (sotah.Profession, error) {
	out := sotah.Profession{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return errors.New("professions bucket was blank")
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
