package sotah

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/statuskinds"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewRegionVersionTimestamps(base64Encoded string) (RegionVersionTimestamps, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return RegionVersionTimestamps{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RegionVersionTimestamps{}, err
	}

	out := RegionVersionTimestamps{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return RegionVersionTimestamps{}, err
	}

	return out, nil
}

// region timestamps

type RegionVersionTimestamps map[blizzardv2.RegionName]VersionRealmTimestamps

func (rvStamps RegionVersionTimestamps) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(rvStamps)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	base64Encoded := base64.StdEncoding.EncodeToString(gzipEncoded)

	return base64Encoded, nil
}

func (rvStamps RegionVersionTimestamps) IsZero() bool {
	for _, vrStamps := range rvStamps {
		for _, rStamps := range vrStamps {
			for _, timestamps := range rStamps {
				if !timestamps.IsZero() {
					return false
				}
			}
		}
	}

	return true
}

func (rvStamps RegionVersionTimestamps) Exists(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) bool {
	vrStamps, ok := rvStamps[tuple.RegionName]
	if !ok {
		return false
	}

	rStamps, ok := vrStamps[tuple.Version]
	if !ok {
		return false
	}

	_, ok = rStamps[tuple.ConnectedRealmId]

	return ok
}

func (rvStamps RegionVersionTimestamps) resolve(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
) RegionVersionTimestamps {
	if _, ok := rvStamps[tuple.RegionName]; !ok {
		rvStamps[tuple.RegionName] = VersionRealmTimestamps{}
	}

	if _, ok := rvStamps[tuple.RegionName][tuple.Version]; !ok {
		rvStamps[tuple.RegionName][tuple.Version] = RealmStatusTimestamps{}
	}

	if _, ok := rvStamps[tuple.RegionName][tuple.Version][tuple.ConnectedRealmId]; !ok {
		rvStamps[tuple.RegionName][tuple.Version][tuple.ConnectedRealmId] = StatusTimestamps{}
	}

	return rvStamps
}

func (rvStamps RegionVersionTimestamps) SetTimestamp(
	tuple blizzardv2.RegionVersionConnectedRealmTuple,
	kind statuskinds.StatusKind,
	timestamp time.Time,
) RegionVersionTimestamps {
	// resolving due to missing members
	out := rvStamps.resolve(tuple)

	// pushing the new time into the found member
	result := out[tuple.RegionName][tuple.Version][tuple.ConnectedRealmId]
	result[kind] = UnixTimestamp(timestamp.Unix())
	out[tuple.RegionName][tuple.Version][tuple.ConnectedRealmId] = result

	return out
}

type VersionRealmTimestamps map[gameversion.GameVersion]RealmStatusTimestamps

type RealmStatusTimestamps map[blizzardv2.ConnectedRealmId]StatusTimestamps
