package sotah

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
)

func NewShortPetList(pets []Pet, locale locale.Locale) []ShortPet {
	out := make([]ShortPet, len(pets))
	for i, pet := range pets {
		out[i] = NewShortPet(ShortPetParams{
			Pet:    pet,
			Locale: locale,
		})
	}

	return out
}

type ShortPetParams struct {
	Pet    Pet
	Locale locale.Locale
}

func NewShortPet(params ShortPetParams) ShortPet {
	return ShortPet{
		Id:   params.Pet.BlizzardMeta.Id,
		Name: params.Pet.BlizzardMeta.Name.FindOr(params.Locale, ""),
	}
}

type ShortPet struct {
	Id      blizzardv2.PetId `json:"id"`
	Name    string           `json:"name"`
	IconUrl string           `json:"icon_url"`
}
