package sotah

import (
	"encoding/json"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
)

func NewStatus(reg Region, stat blizzard.Status) Status {
	return Status{stat, reg, NewRealms(reg, stat.Realms)}
}

type Status struct {
	blizzard.Status
	Region Region `json:"-"`
	Realms Realms `json:"realms"`
}

type Statuses map[blizzard.RegionName]Status

func (s Statuses) RegionRealmsMap() RegionRealmMap {
	out := RegionRealmMap{}

	for regionName, status := range s {
		out[regionName] = RealmMap{}

		for _, realm := range status.Realms {
			out[regionName][realm.Slug] = realm
		}
	}

	return out
}

type Profession struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Icon    string `json:"icon"`
	IconURL string `json:"icon_url"`
}

type Expansion struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Primary    bool   `json:"primary"`
	LabelColor string `json:"label_color"`
}

type UnixTimestamp int64

type WorkerStopChan chan struct{}

func NormalizeTargetDate(targetDate time.Time) time.Time {
	nearestWeekStartOffset := targetDate.Second() + targetDate.Minute()*60 + targetDate.Hour()*60*60
	return time.Unix(targetDate.Unix()-int64(nearestWeekStartOffset), 0)
}

func NewBlizzardCredentials(data []byte) (BlizzardCredentials, error) {
	var out BlizzardCredentials
	if err := json.Unmarshal(data, &out); err != nil {
		return BlizzardCredentials{}, err
	}

	return out, nil
}

type BlizzardCredentials struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (c BlizzardCredentials) EncodeForStorage() ([]byte, error) {
	return json.Marshal(c)
}
