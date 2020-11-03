package professions

import (
	"fmt"
	"strconv"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

// bucketing
func baseBucketName() []byte {
	return []byte("professions")
}

func skillTiersBucketName(professionId blizzardv2.ProfessionId) []byte {
	return []byte(fmt.Sprintf("profession-%d-skill-tiers", professionId))
}

// keying
func baseKeyName(id blizzardv2.ProfessionId) []byte {
	return []byte(fmt.Sprintf("profession-%d", id))
}

func professionIdFromBaseKeyName(key []byte) (blizzardv2.ProfessionId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("profession-"):])
	if err != nil {
		return blizzardv2.ProfessionId(0), err
	}

	return blizzardv2.ProfessionId(unparsedId), nil
}

func skillTiersKeyName(id blizzardv2.SkillTierId) []byte {
	return []byte(fmt.Sprintf("skill-tier-%d", id))
}

func skillTierIdFromSkillTierKeyName(key []byte) (blizzardv2.SkillTierId, error) {
	unparsedId, err := strconv.Atoi(string(key)[len("skill-tier-"):])
	if err != nil {
		return blizzardv2.SkillTierId(0), err
	}

	return blizzardv2.SkillTierId(unparsedId), nil
}

// db
func DatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/professions.db", dbDir), nil
}
