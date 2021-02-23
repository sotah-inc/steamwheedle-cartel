package sotah

import (
	"encoding/json"
	"regexp"
	"strings"
)

type Expansion struct {
	Name       string `json:"name"`
	Label      string `json:"label"`
	Primary    bool   `json:"primary"`
	LabelColor string `json:"label_color"`
}

type WorkerStopChan chan struct{}

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
