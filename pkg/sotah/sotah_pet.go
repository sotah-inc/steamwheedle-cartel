package sotah

import (
	"encoding/json"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
)

func NewPetIconObjectName(name IconName) string {
	return fmt.Sprintf(
		"%s/%s.jpg",
		gameversions.Retail,
		name,
	)
}

type PetIconMeta struct {
	URL        string   `json:"icon_url"`
	ObjectName string   `json:"icon_object_name"`
	Icon       IconName `json:"icon"`
}

type PetMeta struct {
	NormalizedName locale.Mapping `json:"normalized_name"`
	PetIconMeta    PetIconMeta    `json:"pet_icon_meta"`
}

func NewPetFromGzipped(gzipEncoded []byte) (Pet, error) {
	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return Pet{}, err
	}

	return NewPet(gzipDecoded)
}

func NewPet(body []byte) (Pet, error) {
	p := &Pet{}
	if err := json.Unmarshal(body, p); err != nil {
		return Pet{}, err
	}

	return *p, nil
}

type Pet struct {
	BlizzardMeta blizzardv2.PetResponse `json:"blizzard_meta"`
	SotahMeta    PetMeta                `json:"sotah_meta"`
}

func (pet Pet) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(pet)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}
