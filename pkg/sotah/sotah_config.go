package sotah

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

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

type RealmSlugWhitelist map[blizzardv2.RegionName]map[gameversion.GameVersion]blizzardv2.RealmSlugs

func (wl RealmSlugWhitelist) Get(
	regionName blizzardv2.RegionName,
	version gameversion.GameVersion,
) blizzardv2.RealmSlugs {
	versionRealmSlugs, ok := wl[regionName]
	if !ok {
		return blizzardv2.RealmSlugs{}
	}

	realmSlugs, ok := versionRealmSlugs[version]
	if !ok {
		return blizzardv2.RealmSlugs{}
	}

	return realmSlugs
}

type FirebaseConfig struct {
	BrowserApiKey string `json:"browser_api_key"`
}

type Config struct {
	Regions              RegionList                          `json:"regions"`
	GameVersions         gameversion.List                    `json:"game_versions"`
	Whitelist            RealmSlugWhitelist                  `json:"whitelist"`
	UseGCloud            bool                                `json:"use_gcloud"`
	Expansions           []Expansion                         `json:"expansions"`
	PrimarySkillTiers    map[string][]blizzardv2.SkillTierId `json:"primary_skilltiers"`
	ProfessionsBlacklist []blizzardv2.ProfessionId           `json:"professions_blacklist"`
	FirebaseConfig       FirebaseConfig                      `json:"firebase_config"`
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
