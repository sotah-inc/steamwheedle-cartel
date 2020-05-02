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

func (tuples RegionConnectedRealmTuples) FilterByRegionName(name RegionName) RegionConnectedRealmTuples {
	out := RegionConnectedRealmTuples{}
	for _, tuple := range tuples {
		if tuple.RegionName != name {
			continue
		}

		out = append(out, tuple)
	}

	return out
}

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
	LoadConnectedRealmTuple
	RegionHostname string
}

func NewLoadConnectedRealmTuples(data []byte) (LoadConnectedRealmTuples, error) {
	out := LoadConnectedRealmTuples{}
	if err := json.Unmarshal(data, &out); err != nil {
		return LoadConnectedRealmTuples{}, err
	}

	return out, nil
}

type LoadConnectedRealmTuples []LoadConnectedRealmTuple

func (tuples LoadConnectedRealmTuples) RegionConnectedRealmTuples() RegionConnectedRealmTuples {
	out := make(RegionConnectedRealmTuples, len(tuples))
	for i, tuple := range tuples {
		out[i] = tuple.RegionConnectedRealmTuple
	}

	return out
}

func (tuples LoadConnectedRealmTuples) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(tuples)
}

type LoadConnectedRealmTuple struct {
	RegionConnectedRealmTuple
	LastModified time.Time
}

type RegionHostnameTuple struct {
	RegionName     RegionName
	RegionHostname string
}
