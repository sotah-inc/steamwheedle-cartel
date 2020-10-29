package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type ProfessionMeta struct{}

type Profession struct {
	BlizzardMeta blizzardv2.ProfessionResponse `json:"blizzard_meta"`
	SotahMeta    ProfessionMeta                `json:"sotah_meta"`
}

func (profession Profession) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(profession)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}
