package blizzardv2

type LinksBase struct {
	Links SelfReference `json:"_links"`
}

type SelfReference struct {
	Self HrefReference `json:"self"`
}

type HrefReference struct {
	Href string `json:"href"`
}
