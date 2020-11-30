package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortProfessions(
	professions []Profession,
	providedLocale locale.Locale,
) ShortProfessions {
	out := make(ShortProfessions, len(professions))
	for i, profession := range professions {
		out[i] = NewShortProfession(profession, providedLocale)
	}

	return out
}

type ShortProfessions []ShortProfession

type ShortProfessionType struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ShortProfessionSkillTier struct {
	Id        blizzardv2.SkillTierId `json:"id"`
	Name      string                 `json:"name"`
	IsPrimary bool                   `json:"is_primary"`
}

func NewShortProfession(
	profession Profession,
	providedLocale locale.Locale,
) ShortProfession {
	skillTiers := make([]ShortProfessionSkillTier, len(profession.BlizzardMeta.SkillTiers))
	for i, skillTier := range profession.BlizzardMeta.SkillTiers {
		isPrimary := func() bool {
			for _, id := range profession.SotahMeta.PrimarySkillTiers {
				if id == skillTier.Id {
					return true
				}
			}

			return false
		}()

		skillTiers[i] = ShortProfessionSkillTier{
			Id:        skillTier.Id,
			Name:      skillTier.Name.FindOr(providedLocale, ""),
			IsPrimary: isPrimary,
		}
	}

	return ShortProfession{
		Id:          profession.BlizzardMeta.Id,
		Name:        profession.BlizzardMeta.Name.FindOr(providedLocale, ""),
		Description: profession.BlizzardMeta.Description.FindOr(providedLocale, ""),
		Type: ShortProfessionType{
			Type: profession.BlizzardMeta.Type.Type,
			Name: profession.BlizzardMeta.Type.Name.FindOr(providedLocale, ""),
		},
		SkillTiers: skillTiers,
		IconUrl:    profession.SotahMeta.IconUrl,
	}
}

type ShortProfession struct {
	Id          blizzardv2.ProfessionId    `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Type        ShortProfessionType        `json:"type"`
	SkillTiers  []ShortProfessionSkillTier `json:"skilltiers"`
	IconUrl     string                     `json:"icon_url"`
}
