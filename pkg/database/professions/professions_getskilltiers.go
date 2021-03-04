package professions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetSkillTiers(
	professionId blizzardv2.ProfessionId,
	ids []blizzardv2.SkillTierId,
) ([]sotah.SkillTier, error) {
	var out []sotah.SkillTier

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		skillTiersBucket := tx.Bucket(skillTiersBucketName(professionId))
		if skillTiersBucket == nil {
			return errors.New("skill-tiers bucket was blank")
		}

		for _, id := range ids {
			value := skillTiersBucket.Get(skillTiersKeyName(id))
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
