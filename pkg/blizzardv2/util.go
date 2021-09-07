package blizzardv2

import (
	"encoding/json"
	"fmt"
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
	RegionTuple
	Version gameversion.GameVersion `json:"game_version"`
}

func (tuple RegionVersionTuple) Equals(target RegionVersionTuple) bool {
	if tuple.RegionName != target.RegionName {
		return false
	}

	return tuple.Version == target.Version
}

func (tuple RegionVersionTuple) String() string {
	return fmt.Sprintf("%s-%s", tuple.Version, tuple.RegionName)
}

// region/version/connected-realm tuples

type RegionVersionConnectedRealmTuples []RegionVersionConnectedRealmTuple

func (tuples RegionVersionConnectedRealmTuples) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(tuples)
}

func (tuples RegionVersionConnectedRealmTuples) RegionNames() []RegionName {
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

func (tuples RegionVersionConnectedRealmTuples) Flatten() FlatRegionVersionConnectedRealmTuples {
	out := FlatRegionVersionConnectedRealmTuples{}
	for _, tuple := range tuples {
		flatTuple := out.Resolve(tuple.RegionVersionTuple)
		flatTuple.Ids = append(flatTuple.Ids, tuple.ConnectedRealmId)
		out = append(out, flatTuple)
	}

	return out
}

// flattened

type FlatRegionVersionConnectedRealmTuples []FlatRegionVersionConnectedRealmTuple

func (tuples FlatRegionVersionConnectedRealmTuples) Resolve(
	tuple RegionVersionTuple,
) FlatRegionVersionConnectedRealmTuple {
	for _, flatTuple := range tuples {
		if !flatTuple.Tuple.Equals(tuple) {
			continue
		}

		return flatTuple
	}

	return FlatRegionVersionConnectedRealmTuple{
		Tuple: tuple,
		Ids:   []ConnectedRealmId{},
	}
}

type FlatRegionVersionConnectedRealmTuple struct {
	Tuple RegionVersionTuple
	Ids   []ConnectedRealmId
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

func (tuple RegionVersionConnectedRealmTuple) String() string {
	return fmt.Sprintf("%s-%s-%d", tuple.RegionName, tuple.Version, tuple.ConnectedRealmId)
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
	RegionVersionTuple
	RealmSlug RealmSlug `json:"realm_slug"`
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

func (
	tuples LoadConnectedRealmTuples,
) RegionVersionConnectedRealmTuples() RegionVersionConnectedRealmTuples {
	out := make(RegionVersionConnectedRealmTuples, len(tuples))
	for i, tuple := range tuples {
		out[i] = tuple.RegionVersionConnectedRealmTuple
	}

	return out
}

func (tuples LoadConnectedRealmTuples) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(tuples)
}

// load-connected-realm tuple

type LoadConnectedRealmTuple struct {
	RegionVersionConnectedRealmTuple
	LastModified time.Time
}

// region-hostname tuple

type RegionHostnameTuple struct {
	RegionName     RegionName
	RegionHostname string
}

// misc

type PriceValue int64
