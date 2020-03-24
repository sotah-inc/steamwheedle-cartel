package blizzardv2

type ItemId int

type ItemModifier struct {
	Type  int `json:"type"`
	Value int `json:"value"`
}
