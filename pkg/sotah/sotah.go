package sotah

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

type ConfigProfession struct {
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

type UnixTimestamps []UnixTimestamp

func (timestamps UnixTimestamps) Before(limit UnixTimestamp) UnixTimestamps {
	out := UnixTimestamps{}
	for _, timestamp := range timestamps {
		if timestamp > limit {
			continue
		}

		out = append(out, timestamp)
	}

	return out
}

func (timestamps UnixTimestamps) IsZero() bool {
	for _, timestamp := range timestamps {
		if timestamp.IsZero() {
			return false
		}
	}

	return true
}

type UnixTimestamp int64

func (timestamp UnixTimestamp) IsZero() bool {
	return timestamp == 0
}

func (timestamp UnixTimestamp) Time() time.Time {
	return time.Unix(int64(timestamp), 0)
}

type WorkerStopChan chan struct{}

func NormalizeToWeek(targetDate time.Time) time.Time {
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

func NormalizeString(input string) (string, error) {
	reg, err := regexp.Compile("[^a-z0-9 ]+")
	if err != nil {
		return "", err
	}

	return reg.ReplaceAllString(strings.ToLower(input), ""), nil
}
