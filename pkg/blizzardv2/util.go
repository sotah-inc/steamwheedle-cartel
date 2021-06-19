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

type RegionTuple struct {
	RegionName RegionName `json:"region_name"`
}

// region/version tuple

type RegionVersionTuple struct {
	RegionTuple
	Version gameversion.GameVersion `json:"game_version"`
}

// region/version/connected-realm tuples

type RegionVersionConnectedRealmTuples []RegionVersionConnectedRealmTuple

func (tuples RegionVersionConnectedRealmTuples) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(tuples)
}

// region/version/connected-realm tuple

type RegionVersionConnectedRealmTuple struct {
	RegionVersionTuple
	ConnectedRealmId ConnectedRealmId `json:"connected_realm_id"`
}

func (tuple RegionVersionConnectedRealmTuple) String() string {
	return fmt.Sprintf("%s-%s-%d", tuple.RegionName, tuple.Version, tuple.ConnectedRealmId)
}

// region/version/realm tuple

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
