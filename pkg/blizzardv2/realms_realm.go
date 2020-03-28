package blizzardv2

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

type RealmId int

type RealmSlug string

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
		Type string         `json:"type"`
		Name locale.Mapping `json:"string"`
	} `json:"type"`
	IsTournament bool      `json:"is_tournament"`
	Slug         RealmSlug `json:"slug"`
}
