package blizzard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

// NewRealmFromFilepath loads a realm from a json file
func NewRealmFromFilepath(relativeFilepath string) (Realm, error) {
	body, err := util.ReadFile(relativeFilepath)
	if err != nil {
		return Realm{}, err
	}

	return newRealm(body)
}

func newRealm(body []byte) (Realm, error) {
	rea := &Realm{}
	if err := json.Unmarshal(body, &rea); err != nil {
		return Realm{}, err
	}

	return *rea, nil
}

// RealmSlug is the region-specific unique identifier
type RealmSlug string

type RealmKey struct {
	Href string `json:"href"`
}

type RealmId int

// Realm represents a given realm
type Realm struct {
	Key  RealmKey  `json:"key"`
	Name string    `json:"name"`
	Id   RealmId   `json:"id"`
	Slug RealmSlug `json:"slug"`
}

const statusURLFormat = "https://%s/data/wow/realm/index?namespace=dynamic-%s&locale=en_US"

// GetStatusURLFunc defines the expected func signature for generating a status uri
type GetStatusURLFunc func(string, string) string

// DefaultGetStatusURL returns a formatted uri
func DefaultGetStatusURL(regionHostname string, regionName string) string {
	return fmt.Sprintf(statusURLFormat, regionHostname, regionName)
}

// NewStatusFromHTTP loads a status from a uri
func NewStatusFromHTTP(uri string) (Status, ResponseMeta, error) {
	resp, err := Download(uri)
	if err != nil {
		return Status{}, resp, err
	}

	if resp.Status != http.StatusOK {
		return Status{}, resp, errors.New("status was not 200")
	}

	status, err := NewStatus(resp.Body)
	if err != nil {
		return Status{}, resp, err
	}

	return status, resp, nil
}

// NewStatusFromFilepath loads a status from a json file
func NewStatusFromFilepath(relativeFilepath string) (Status, error) {
	body, err := util.ReadFile(relativeFilepath)
	if err != nil {
		return Status{}, err
	}

	return NewStatus(body)
}

// NewStatus loads a status from a byte array of json
func NewStatus(body []byte) (Status, error) {
	s := &Status{}
	if err := json.Unmarshal(body, s); err != nil {
		return Status{}, err
	}

	return *s, nil
}

// Status contains a list of realms
type Status struct {
	Realms []Realm `json:"realms"`
}
