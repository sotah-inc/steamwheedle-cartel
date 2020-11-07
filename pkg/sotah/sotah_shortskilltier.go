package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortSkillTier(skillTier SkillTier, providedLocale locale.Locale) ShortSkillTier {
	return ShortSkillTier{
		Id: skillTier.BlizzardMeta.Id,
	}
}

type ShortSkillTier struct {
	Id blizzardv2.SkillTierId `json:"id"`
}
