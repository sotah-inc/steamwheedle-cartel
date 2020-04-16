package blizzardv2

import (
	"encoding/json"
	"time"
)

type LinksBase struct {
	Links SelfReference `json:"_links"`
}

type SelfReference struct {
	Self HrefReference `json:"self"`
}

type HrefReference struct {
	Href string `json:"href"`
}

func NewRegionConnectedRealmTuples(data []byte) (RegionConnectedRealmTuples, error) {
	out := RegionConnectedRealmTuples{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionConnectedRealmTuples{}, err
	}

	return out, nil
}

type RegionConnectedRealmTuples []RegionConnectedRealmTuple

func NewRegionConnectedRealmTuple(data []byte) (RegionConnectedRealmTuple, error) {
	out := RegionConnectedRealmTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionConnectedRealmTuple{}, err
	}

	return out, nil
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
