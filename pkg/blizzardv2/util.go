package blizzardv2

import "time"

type LinksBase struct {
	Links SelfReference `json:"_links"`
}

type SelfReference struct {
	Self HrefReference `json:"self"`
}

type HrefReference struct {
	Href string `json:"href"`
}

type RegionConnectedRealmTuple struct {
	RegionName       RegionName
	ConnectedRealmId ConnectedRealmId
}

type DownloadConnectedRealmTuple struct {
	RegionConnectedRealmTuple
	RegionHostname string
	LastModified   time.Time
}
