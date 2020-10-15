package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const petURLFormat = "https://%s/data/wow/pet/%d?namespace=static-%s"

func DefaultPetURL(regionHostname string, regionName RegionName, id PetId) string {
	return fmt.Sprintf(petURLFormat, regionHostname, id, regionName)
}

type GetPetURLFunc func(string) string

type PetId int

type PetAbility struct {
	Ability struct {
		Key  HrefReference  `json:"key"`
		Name locale.Mapping `json:"name"`
	} `json:"ability"`
	Slot          int `json:"slot"`
	RequiredLevel int `json:"required_level"`
}

type PetResponse struct {
	LinksBase
	Id            PetId          `json:"id"`
	Name          locale.Mapping `json:"name"`
	BattlePetType struct {
		Id   int            `json:"id"`
		Type string         `json:"type"`
		Name locale.Mapping `json:"name"`
	} `json:"battle_pet_type"`
	Description    locale.Mapping `json:"description"`
	IsCapturable   bool           `json:"is_capturable"`
	IsTradable     bool           `json:"is_tradable"`
	IsBattlepet    bool           `json:"is_battlepet"`
	IsAllianceOnly bool           `json:"is_alliance_only"`
	IsHordeOnly    bool           `json:"is_horde_only"`
	Abilities      []PetAbility   `json:"abilities"`
	Source         struct {
		Type string         `json:"type"`
		Name locale.Mapping `json:"name"`
	} `json:"source"`
	Icon     string `json:"icon"`
	Creature struct {
		Key  HrefReference  `json:"key"`
		Name locale.Mapping `json:"name"`
		Id   int            `json:"id"`
	} `json:"creature"`
	IsRandomCreatureDisplay bool `json:"is_random_creature_display"`
	Media                   struct {
		Key HrefReference `json:"key"`
		Id  PetId         `json:"id"`
	} `json:"media"`
}

func NewPetResponse(body []byte) (PetResponse, error) {
	pet := &PetResponse{}
	if err := json.Unmarshal(body, pet); err != nil {
		return PetResponse{}, err
	}

	return *pet, nil
}

func NewPetFromHTTP(uri string) (PetResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download pet")

		return PetResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from pet was not 200")

		return PetResponse{}, resp, errors.New("status was not 200")
	}

	pet, err := NewPetResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse pet response")

		return PetResponse{}, resp, err
	}

	return pet, resp, nil
}
