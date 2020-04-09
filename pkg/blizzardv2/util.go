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
	RegionName       RegionName       `json:"region_name"`
	ConnectedRealmId ConnectedRealmId `json:"connected_realm_id"`
}

type DownloadConnectedRealmTuple struct {
	RegionConnectedRealmTuple
	RegionHostname string
	LastModified   time.Time
}

type RegionHostnameTuple struct {
	RegionName     RegionName
	RegionHostname string
}
