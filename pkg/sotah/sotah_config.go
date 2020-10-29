package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewConfigFromFilepath(relativePath string) (Config, error) {
	logging.WithField("path", relativePath).Info("reading config")

	body, err := util.ReadFile(relativePath)
	if err != nil {
		return Config{}, err
	}

	return NewConfig(body)
}

func NewConfig(body []byte) (Config, error) {
	c := &Config{}
	if err := json.Unmarshal(body, &c); err != nil {
		return Config{}, err
	}

	return *c, nil
}

type RegionRealmSlugWhitelist map[blizzardv2.RegionName]blizzardv2.RealmSlugs

func (wl RegionRealmSlugWhitelist) Get(name blizzardv2.RegionName) blizzardv2.RealmSlugs {
	realmSlugs, ok := wl[name]
	if !ok {
		return blizzardv2.RealmSlugs{}
	}

	return realmSlugs
}

type Config struct {
	Regions     RegionList               `json:"regions"`
	Whitelist   RegionRealmSlugWhitelist `json:"whitelist"`
	UseGCloud   bool                     `json:"use_gcloud"`
	Expansions  []Expansion              `json:"expansions"`
	Professions []ConfigProfession       `json:"professions"`
}

func (c Config) FilterInRegions(regs RegionList) RegionList {
	out := RegionList{}

	for _, reg := range regs {
		if _, ok := c.Whitelist[reg.Name]; !ok {
			continue
		}

		out = append(out, reg)
	}

	return out
}
