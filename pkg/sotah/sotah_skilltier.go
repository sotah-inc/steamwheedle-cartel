package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type SkillTierMeta struct{}

type SkillTier struct {
	BlizzardMeta blizzardv2.SkillTierResponse `json:"blizzard_meta"`
	SotahMeta    SkillTierMeta                `json:"sotah_meta"`
}

func (SkillTier SkillTier) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(SkillTier)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}
