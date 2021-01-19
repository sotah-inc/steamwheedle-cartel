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

type WorkerStopChan chan struct{}

func NormalizeToDay(targetTimestamp UnixTimestamp) UnixTimestamp {
	targetDate := time.Unix(int64(targetTimestamp), 0)
	nearestDayStartOffset := targetDate.Second() + targetDate.Minute()*60 + targetDate.Hour()*60*60
	return UnixTimestamp(targetDate.Unix() - int64(nearestDayStartOffset))
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
