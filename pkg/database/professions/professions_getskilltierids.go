package professions

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (pdBase Database) GetSkillTierIds(professionId blizzardv2.ProfessionId) ([]blizzardv2.SkillTierId, error) {
	var out []blizzardv2.SkillTierId

	// peeking into the items database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		stBucket := tx.Bucket(skillTiersBucketName(professionId))
		if stBucket == nil {
			return nil
		}

		return stBucket.ForEach(func(k []byte, v []byte) error {
			id, err := skillTierIdFromKeyName(k)
			if err != nil {
				return err
			}

			out = append(out, id)

			return nil
		})
	})
	if err != nil {
		return []blizzardv2.SkillTierId{}, err
	}

	return out, nil
}
