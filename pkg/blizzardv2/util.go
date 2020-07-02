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

func NewRegionTuple(data []byte) (RegionTuple, error) {
	out := RegionTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionTuple{}, err
	}

	return out, nil
}

type RegionTuple struct {
	RegionName RegionName `json:"region_name"`
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

func (tuples RegionConnectedRealmTuples) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(tuples)
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

func NewRegionRealmTuple(data []byte) (RegionRealmTuple, error) {
	out := RegionRealmTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionRealmTuple{}, err
	}

	return out, nil
}

type RegionRealmTuple struct {
	RegionName RegionName `json:"region_name"`
	RealmSlug  RealmSlug  `json:"realm_slug"`
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

type PriceValue int64
