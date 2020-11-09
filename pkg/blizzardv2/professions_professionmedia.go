package blizzardv2

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type ProfessionMediaResponseAsset struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	FileDataId int    `json:"file_data_id"`
}

type ProfessionMediaResponse struct {
	LinksBase
	Assets []ProfessionMediaResponseAsset `json:"assets"`
	Id     ProfessionId                   `json:"id"`
}

func (res ProfessionMediaResponse) GetIconUrl() (string, error) {
	if len(res.Assets) == 0 {
		return "", errors.New("assets was blank")
	}

	return res.Assets[0].Value, nil
}

func NewProfessionMediaResponseFromHTTP(uri string) (ProfessionMediaResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download profession-media")

		return ProfessionMediaResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from profession-media was not 200")

		return ProfessionMediaResponse{}, resp, errors.New("status was not 200")
	}

	professionMedia, err := NewProfessionMediaResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse profession-media response")

		return ProfessionMediaResponse{}, resp, err
	}

	return professionMedia, resp, nil
}

func NewProfessionMediaResponse(body []byte) (ProfessionMediaResponse, error) {
	professionMedia := &ProfessionMediaResponse{}
	if err := json.Unmarshal(body, professionMedia); err != nil {
		return ProfessionMediaResponse{}, err
	}

	return *professionMedia, nil
}
