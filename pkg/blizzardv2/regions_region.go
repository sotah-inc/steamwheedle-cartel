package blizzardv2

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

type RegionId int

type RegionName string

type Region struct {
	LinksBase
	Id   RegionId       `json:"id"`
	Name locale.Mapping `json:"name"`
	Tag  string         `json:"tag"`
}
