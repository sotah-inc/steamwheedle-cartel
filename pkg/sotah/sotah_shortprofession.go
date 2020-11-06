package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortProfessions(professions []Profession, providedLocale locale.Locale) ShortProfessions {
	out := make(ShortProfessions, len(professions))
	for i, profession := range professions {
		out[i] = NewShortProfession(profession, providedLocale)
	}

	return out
}

type ShortProfessions []ShortProfession

func NewShortProfession(profession Profession, providedLocale locale.Locale) ShortProfession {
	return ShortProfession{
		Id:   profession.BlizzardMeta.Id,
		Name: profession.BlizzardMeta.Name.FindOr(providedLocale, ""),
	}
}

type ShortProfession struct {
	Id   blizzardv2.ProfessionId `json:"id"`
	Name string                  `json:"name"`
}
