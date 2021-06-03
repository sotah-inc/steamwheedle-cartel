package blizzardv2

import (
	"encoding/json"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
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

func NewVersionRegionTuple(data []byte) (VersionRegionTuple, error) {
	out := VersionRegionTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return VersionRegionTuple{}, err
	}

	return out, nil
}

type VersionRegionTuple struct {
	Version    gameversion.GameVersion `json:"game_version"`
	RegionName RegionName              `json:"region_name"`
}

func NewRegionConnectedRealmTuples(data []byte) (RegionConnectedRealmTuples, error) {
	out := RegionConnectedRealmTuples{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionConnectedRealmTuples{}, err
	}

	return out, nil
}

type RegionConnectedRealmTuples []RegionConnectedRealmTuple

func (tuples RegionConnectedRealmTuples) FilterByRegionName(
	name RegionName,
) RegionConnectedRealmTuples {
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

func (tuples RegionConnectedRealmTuples) RegionNames() []RegionName {
	outMap := map[RegionName]struct{}{}
	for _, tuple := range tuples {
		outMap[tuple.RegionName] = struct{}{}
	}

	out := make([]RegionName, len(outMap))
	i := 0
	for regionName := range outMap {
		out[i] = regionName
		i += 1
	}

	return out
}

func (tuples RegionConnectedRealmTuples) ToMap() map[RegionName][]ConnectedRealmId {
	out := map[RegionName][]ConnectedRealmId{}
	for _, name := range tuples.RegionNames() {
		out[name] = []ConnectedRealmId{}
	}

	for _, tuple := range tuples {
		ids := out[tuple.RegionName]
		ids = append(ids, tuple.ConnectedRealmId)

		out[tuple.RegionName] = ids
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

func NewVersionRegionConnectedRealmTuple(data []byte) (VersionRegionConnectedRealmTuple, error) {
	out := VersionRegionConnectedRealmTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return VersionRegionConnectedRealmTuple{}, err
	}

	return out, nil
}

type VersionRegionConnectedRealmTuple struct {
	Version          gameversion.GameVersion `json:"game_version"`
	RegionName       RegionName              `json:"region_name"`
	ConnectedRealmId ConnectedRealmId        `json:"connected_realm_id"`
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

func NewVersionRegionRealmTuple(data []byte) (VersionRegionRealmTuple, error) {
	out := VersionRegionRealmTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return VersionRegionRealmTuple{}, err
	}

	return out, nil
}

type VersionRegionRealmTuple struct {
	Version    gameversion.GameVersion `json:"game_version"`
	RegionName RegionName              `json:"region_name"`
	RealmSlug  RealmSlug               `json:"realm_slug"`
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

func (tuples LoadConnectedRealmTuples) RegionNames() []RegionName {
	outMap := map[RegionName]struct{}{}
	for _, tuple := range tuples {
		outMap[tuple.RegionName] = struct{}{}
	}

	out := make([]RegionName, len(outMap))
	i := 0
	for name := range outMap {
		out[i] = name

		i += 1
	}

	return out
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
