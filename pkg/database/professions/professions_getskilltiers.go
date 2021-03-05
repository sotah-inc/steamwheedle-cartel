package professions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetSkillTiers(
	tuples []blizzardv2.ProfessionSkillTierTuple,
) ([]sotah.SkillTier, error) {
	var out []sotah.SkillTier

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		for _, tuple := range tuples {
			skillTiersBucket := tx.Bucket(skillTiersBucketName(tuple.ProfessionId))
			if skillTiersBucket == nil {
				return errors.New("skill-tiers bucket was blank")
			}

			value := skillTiersBucket.Get(skillTiersKeyName(tuple.SkillTierId))
			if value == nil {
				continue
			}

			skillTier, err := sotah.NewSkillTier(value)
			if err != nil {
				return err
			}

			out = append(out, skillTier)
		}

		return nil
	})
	if err != nil {
		return []sotah.SkillTier{}, err
	}

	return out, nil
}
