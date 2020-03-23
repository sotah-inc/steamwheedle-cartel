package blizzardv2

type ConnectedRealmId int

type ConnectedRealmResponse struct {
	LinksBase
	Id       ConnectedRealmId `json:"id"`
	HasQueue bool             `json:"has_queue"`
	Status   struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"status"`
	Population struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"population"`

	MythicLeaderboards HrefReference `json:"mythic_leaderboards"`
	Auctions           HrefReference `json:"auctions"`
}
