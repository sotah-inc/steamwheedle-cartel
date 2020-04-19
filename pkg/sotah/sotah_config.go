package sotah

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewConfigFromFilepath(relativePath string) (Config, error) {
	logging.WithField("path", relativePath).Info("Reading Config")

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

type RegionRealmSlugWhitelist map[blizzardv2.RegionName][]blizzardv2.RealmSlug

func (wl RegionRealmSlugWhitelist) Has(name blizzardv2.RegionName, res blizzardv2.ConnectedRealmResponse) bool {
	realmSlugs, ok := wl[name]
	if !ok {
		logging.WithFields(logrus.Fields{
			"region":                   name,
			"whitelist":                wl,
			"connected-realm-response": res,
		}).Info("whitelist did not have region")

		return false
	}

	for _, realm := range res.Realms {
		for _, whitelistSlug := range realmSlugs {
			if realm.Slug == whitelistSlug {
				logging.WithFields(logrus.Fields{
					"region":                   name,
					"realm":                    realm.Slug,
					"whitelist":                wl,
					"connected-realm-response": res,
				}).Info("found realm in whitelist")

				return true
			}
		}
	}

	logging.WithFields(logrus.Fields{
		"region":                   name,
		"whitelist":                wl,
		"connected-realm-response": res,
	}).Info("whitelist did have region but response did not contain realm")

	return false
}

type Config struct {
	Regions       RegionList               `json:"regions"`
	Whitelist     RegionRealmSlugWhitelist `json:"whitelist"`
	UseGCloud     bool                     `json:"use_gcloud"`
	Expansions    []Expansion              `json:"expansions"`
	Professions   []Profession             `json:"professions"`
	ItemBlacklist []blizzardv2.ItemId      `json:"item_blacklist"`
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
