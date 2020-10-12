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

const petIndexURLFormat = "https://%s/data/wow/pet/index?namespace=static-%s"

func DefaultPetIndexURL(regionHostname string, regionName RegionName) string {
	return fmt.Sprintf(petIndexURLFormat, regionHostname, regionName)
}

type GetPetIndexURLFunc func(string) string

type PetIndexPet struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   PetId          `json:"id"`
}

type PetIndexResponse struct {
	LinksBase
	Pets []PetIndexPet `json:"pets"`
}

func NewPetIndexResponse(body []byte) (PetIndexResponse, error) {
	pIndex := &PetIndexResponse{}
	if err := json.Unmarshal(body, pIndex); err != nil {
		return PetIndexResponse{}, err
	}

	return *pIndex, nil
}

func NewPetIndexFromHTTP(uri string) (PetIndexResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download pet-index")

		return PetIndexResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from pet-index was not 200")

		return PetIndexResponse{}, resp, errors.New("status was not 200")
	}

	pIndex, err := NewPetIndexResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse pet-index response")

		return PetIndexResponse{}, resp, err
	}

	return pIndex, resp, nil
}
