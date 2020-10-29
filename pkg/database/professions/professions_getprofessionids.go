package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (pdBase Database) GetProfessionIds() ([]blizzardv2.ProfessionId, error) {
	var out []blizzardv2.ProfessionId

	// peeking into the items database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		baseBucket := tx.Bucket(baseBucketName())
		if baseBucket == nil {
			return nil
		}

		return baseBucket.ForEach(func(k []byte, v []byte) error {
			id, err := professionIdFromBaseKeyName(k)
			if err != nil {
				return err
			}

			out = append(out, id)

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.ProfessionId{}, err
	}

	return out, nil
}
