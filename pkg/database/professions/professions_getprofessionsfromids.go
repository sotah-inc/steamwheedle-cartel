package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetProfessionsFromIds(
	ids []blizzardv2.ProfessionId,
) ([]sotah.Profession, error) {
	var out []sotah.Profession

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return nil
		}

		for _, id := range ids {
			value := baseBucket.Get(baseKeyName(id))
			if value == nil {
				continue
			}

			profession, err := sotah.NewProfession(value)
			if err != nil {
				return err
			}

			out = append(out, profession)
		}

		return nil

	})
	if err != nil {
		return []sotah.Profession{}, err
	}

	return out, nil
}
