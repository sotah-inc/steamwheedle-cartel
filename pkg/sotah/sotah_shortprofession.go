package sotah

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

func NewShortProfessions(professions []Profession) ShortProfessions {
	out := make(ShortProfessions, len(professions))
	for i, profession := range professions {
		out[i] = NewShortProfession(profession)
	}

	return out
}

type ShortProfessions []ShortProfession

func NewShortProfession(profession Profession) ShortProfession {
	return ShortProfession{
		Id: profession.BlizzardMeta.Id,
	}
}

type ShortProfession struct {
	Id blizzardv2.ProfessionId `json:"id"`
}
