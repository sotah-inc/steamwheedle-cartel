package blizzardv2

type RegionId int

type Region struct {
	LinksBase
	Id   RegionId `json:"id"`
	Name string   `json:"name"`
	Tag  string   `json:"tag"`
}
