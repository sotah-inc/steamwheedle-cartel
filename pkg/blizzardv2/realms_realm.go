package blizzardv2

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/realmtype"
)

type RealmId int

type RealmSlug string

type RealmSlugs []RealmSlug

func (slugs RealmSlugs) Has(providedSlug RealmSlug) bool {
	for _, slug := range slugs {
		if slug == providedSlug {
			return true
		}
	}

	return false
}

type RealmResponse struct {
	LinksBase
	Id     RealmId `json:"id"`
	Region struct {
		LinksBase
		Id   RegionId       `json:"id"`
		Name locale.Mapping `json:"name"`
	} `json:"region"`
	ConnectedRealm HrefReference  `json:"connected_realm"`
	Name           locale.Mapping `json:"name"`
	Category       locale.Mapping `json:"category"`
	Locale         string         `json:"locale"`
	Timezone       string         `json:"timezone"`
	Type           struct {
		Type realmtype.RealmType `json:"type"`
		Name locale.Mapping      `json:"name"`
	} `json:"type"`
	IsTournament bool      `json:"is_tournament"`
	Slug         RealmSlug `json:"slug"`
}

type RealmResponses []RealmResponse

func (responses RealmResponses) FilterIn(wl RealmSlugs) RealmResponses {
	out := RealmResponses{}
	for _, res := range responses {
		if !wl.Has(res.Slug) {
			continue
		}

		out = append(out, res)
	}

	return out
}
