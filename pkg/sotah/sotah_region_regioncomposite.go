package sotah

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

// realm-composites

type RealmComposites []RealmComposite

func (comps RealmComposites) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(comps)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

// realm-composite

func NewRealmCompositeFromStorage(
	data []byte,
) (RealmComposite, error) {
	out := RealmComposite{}
	if err := json.Unmarshal(data, &out); err != nil {
		return RealmComposite{}, err
	}

	return out, nil
}

type RealmComposite struct {
	ConnectedRealmResponse blizzardv2.ConnectedRealmResponse `json:"connected_realm"`
	StatusTimestamps       StatusTimestamps                  `json:"status_timestamps"`
}

func (composite RealmComposite) IsZero() bool {
	return len(composite.ConnectedRealmResponse.Realms) == 0
}

func (composite RealmComposite) EncodeForStorage() ([]byte, error) {
	return json.Marshal(composite)
}

// version-realm composites

type VersionRealmComposites map[gameversion.GameVersion]RealmComposites

// region-composite

type RegionComposite struct {
	ConfigRegion             Region                 `json:"config_region"`
	ConnectedRealmComposites VersionRealmComposites `json:"connected_realms"`
}

func (region RegionComposite) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(region)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}
