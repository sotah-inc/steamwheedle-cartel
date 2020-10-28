package blizzardv2

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

type ProfessionSkillTierId int

type ProfessionSkillTier struct {
	Key  HrefReference         `json:"key"`
	Name locale.Mapping        `json:"name"`
	Id   ProfessionSkillTierId `json:"id"`
}

type ProfessionId int

type Profession struct {
	LinksBase
	Id          ProfessionId   `json:"id"`
	Name        locale.Mapping `json:"name"`
	Description locale.Mapping `json:"description"`
	Type        struct {
		Type string         `json:"type"`
		Name locale.Mapping `json:"name"`
	} `json:"type"`
	Media struct {
		Key HrefReference `json:"key"`
		Id  ProfessionId  `json:"id"`
	} `json:"media"`
	SkillTiers []ProfessionSkillTier `json:"skill_tiers"`
}
