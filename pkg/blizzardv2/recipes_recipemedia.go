package blizzardv2

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type RecipeMediaResponseAsset struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	FileDataId int    `json:"file_data_id"`
}

type RecipeMediaResponse struct {
	LinksBase
	Assets []RecipeMediaResponseAsset `json:"assets"`
	Id     RecipeId                   `json:"id"`
}

func (res RecipeMediaResponse) GetIconUrl() (string, error) {
	if len(res.Assets) == 0 {
		return "", errors.New("assets was blank")
	}

	return res.Assets[0].Value, nil
}

func NewRecipeMediaResponse(body []byte) (RecipeMediaResponse, error) {
	rmResp := &RecipeMediaResponse{}
	if err := json.Unmarshal(body, rmResp); err != nil {
		return RecipeMediaResponse{}, err
	}

	return *rmResp, nil
}

func NewRecipeMediaResponseFromHTTP(uri string) (RecipeMediaResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download recipe-media")

		return RecipeMediaResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from recipe-media was not 200")

		return RecipeMediaResponse{}, resp, errors.New("status was not 200")
	}

	rmResp, err := NewRecipeMediaResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse recipe-media response")

		return RecipeMediaResponse{}, resp, err
	}

	return rmResp, resp, nil
}
