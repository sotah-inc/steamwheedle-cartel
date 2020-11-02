package blizzardv2

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type ProfessionResponseSkillTier struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   SkillTierId    `json:"id"`
}

type ProfessionId int

type ProfessionResponse struct {
	LinksBase
	Id          ProfessionId   `json:"id"`
	Name        locale.Mapping `json:"name"`
	Description locale.Mapping `json:"description"`
	Type        struct {
		Type string         `json:"type"`
		Name locale.Mapping `json:"name"`
	} `json:"type"`
	Media struct {
		Key HrefReference `json:"key"`
		Id  ProfessionId  `json:"id"`
	} `json:"media"`
	SkillTiers []ProfessionResponseSkillTier `json:"skill_tiers"`
}

func NewProfessionResponseFromHTTP(uri string) (ProfessionResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download profession")

		return ProfessionResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from profession was not 200")

		return ProfessionResponse{}, resp, errors.New("status was not 200")
	}

	profession, err := NewProfessionResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse profession response")

		return ProfessionResponse{}, resp, err
	}

	return profession, resp, nil
}

func NewProfessionResponse(body []byte) (ProfessionResponse, error) {
	profession := &ProfessionResponse{}
	if err := json.Unmarshal(body, profession); err != nil {
		return ProfessionResponse{}, err
	}

	return *profession, nil
}
