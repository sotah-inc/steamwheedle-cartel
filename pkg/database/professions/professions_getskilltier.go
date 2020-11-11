package professions

import (
	"errors"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (pdBase Database) GetSkillTier(
	professionId blizzardv2.ProfessionId,
	skillTierId blizzardv2.SkillTierId,
) (sotah.SkillTier, error) {
	out := sotah.SkillTier{}

	// peeking into the professions database
	err := pdBase.db.View(func(tx *bolt.Tx) error {
		skillTiersBucket := tx.Bucket(skillTiersBucketName(professionId))
		if skillTiersBucket == nil {
			return errors.New("skill-tiers bucket was blank")
		}

		data := skillTiersBucket.Get(skillTiersKeyName(skillTierId))
		if data == nil {
			return errors.New("skill-tier not found")
		}

		var err error
		out, err = sotah.NewSkillTier(data)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.SkillTier{}, err
	}

	return out, nil
}
