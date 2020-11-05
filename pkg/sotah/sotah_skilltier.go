package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type SkillTierMeta struct{}

func NewSkillTier(body []byte) (SkillTier, error) {
	gzipDecoded, err := util.GzipDecode(body)
	if err != nil {
		return SkillTier{}, err
	}

	out := SkillTier{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return SkillTier{}, err
	}

	return out, nil
}

type SkillTier struct {
	BlizzardMeta blizzardv2.SkillTierResponse `json:"blizzard_meta"`
	SotahMeta    SkillTierMeta                `json:"sotah_meta"`
}

func (skillTier SkillTier) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(skillTier)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

func (skillTier SkillTier) RecipeIds() []blizzardv2.RecipeId {
	var out []blizzardv2.RecipeId

	for _, category := range skillTier.BlizzardMeta.Categories {
		for _, recipe := range category.Recipes {
			out = append(out, recipe.Id)
		}
	}

	return out
}

func NewSkillTiersIntakeRequest(body []byte) (SkillTiersIntakeRequest, error) {
	out := &SkillTiersIntakeRequest{}
	if err := json.Unmarshal(body, out); err != nil {
		return SkillTiersIntakeRequest{}, err
	}

	return *out, nil
}

type SkillTiersIntakeRequest struct {
	ProfessionId blizzardv2.ProfessionId `json:"profession_id"`
}

func (req SkillTiersIntakeRequest) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(req)
}
