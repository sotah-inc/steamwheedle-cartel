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

// region tuple

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

// region/version tuple

func NewRegionVersionTuple(data []byte) (RegionVersionTuple, error) {
	out := RegionVersionTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionVersionTuple{}, err
	}

	return out, nil
}

type RegionVersionTuple struct {
	RegionName RegionName              `json:"region_name"`
	Version    gameversion.GameVersion `json:"game_version"`
}

// region/connected-realm tuples

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

// region/connected-realm tuple

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

// region/version/connected-realm tuple

func NewRegionVersionConnectedRealmTuple(data []byte) (RegionVersionConnectedRealmTuple, error) {
	out := RegionVersionConnectedRealmTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionVersionConnectedRealmTuple{}, err
	}

	return out, nil
}

type RegionVersionConnectedRealmTuple struct {
	RegionVersionTuple
	ConnectedRealmId ConnectedRealmId `json:"connected_realm_id"`
}

// region/realm tuple

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

// region/version/realm tuple

func NewRegionVersionRealmTuple(data []byte) (RegionVersionRealmTuple, error) {
	out := RegionVersionRealmTuple{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RegionVersionRealmTuple{}, err
	}

	return out, nil
}

type RegionVersionRealmTuple struct {
	RegionRealmTuple
	Version gameversion.GameVersion `json:"game_version"`
}

// download-connected-realm tuple

type DownloadConnectedRealmTuple struct {
	LoadConnectedRealmTuple
	RegionHostname string
}

// load-connected-realm tuples

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

// load-connected-realm tuple

type LoadConnectedRealmTuple struct {
	RegionConnectedRealmTuple
	LastModified time.Time
}

// region-hostname tuple

type RegionHostnameTuple struct {
	RegionName     RegionName
	RegionHostname string
}

// misc

type PriceValue int64
