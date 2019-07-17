package bus

import (
	"encoding/json"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func NewLoadRegionRealmTimestampsInJob(data string) (LoadRegionRealmTimestampsInJob, error) {
	var out LoadRegionRealmTimestampsInJob
	if err := json.Unmarshal([]byte(data), &out); err != nil {
		return LoadRegionRealmTimestampsInJob{}, err
	}

	return out, nil
}

type LoadRegionRealmTimestampsInJob struct {
	RegionName      string `json:"region_name"`
	RealmSlug       string `json:"realm_slug"`
	TargetTimestamp int    `json:"target_timestamp"`
}

func (j LoadRegionRealmTimestampsInJob) EncodeForDelivery() (string, error) {
	out, err := json.Marshal(j)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func (j LoadRegionRealmTimestampsInJob) ToRegionRealmTimestampTuple() RegionRealmTimestampTuple {
	return RegionRealmTimestampTuple{
		RegionName:      j.RegionName,
		RealmSlug:       j.RealmSlug,
		TargetTimestamp: j.TargetTimestamp,
	}
}

func (j LoadRegionRealmTimestampsInJob) ToRealmTime() (sotah.Realm, time.Time) {
	realm := sotah.NewSkeletonRealm(blizzard.RegionName(j.RegionName), blizzard.RealmSlug(j.RealmSlug))
	targetTime := time.Unix(int64(j.TargetTimestamp), 0)

	return realm, targetTime
}
