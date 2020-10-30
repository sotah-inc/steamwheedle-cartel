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

const professionIndexURLFormat = "https://%s/data/wow/profession/index?namespace=static-%s"

func DefaultProfessionIndexURL(regionHostname string, regionName RegionName) string {
	return fmt.Sprintf(professionIndexURLFormat, regionHostname, regionName)
}

type GetProfessionIndexURLFunc func(string) string

type ProfessionsIndexProfession struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   ProfessionId   `json:"id"`
}

type ProfessionsIndexResponse struct {
	LinksBase

	Professions []ProfessionsIndexProfession `json:"professions"`
}

func NewProfessionsIndexResponseFromHTTP(uri string) (ProfessionsIndexResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download professions-index")

		return ProfessionsIndexResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from professions-index was not 200")

		return ProfessionsIndexResponse{}, resp, errors.New("status was not 200")
	}

	pIndex, err := NewProfessionsIndexResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse professions-index response")

		return ProfessionsIndexResponse{}, resp, err
	}

	return pIndex, resp, nil
}

func NewProfessionsIndexResponse(body []byte) (ProfessionsIndexResponse, error) {
	pIndex := &ProfessionsIndexResponse{}
	if err := json.Unmarshal(body, pIndex); err != nil {
		return ProfessionsIndexResponse{}, err
	}

	return *pIndex, nil
}
