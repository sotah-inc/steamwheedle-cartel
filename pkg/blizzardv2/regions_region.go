package blizzardv2

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type RegionId int

type RegionName string

type RegionResponse struct {
	LinksBase
	Id   RegionId       `json:"id"`
	Name locale.Mapping `json:"name"`
	Tag  string         `json:"tag"`
}

func NewRegionResponseFromHTTP(uri string) (RegionResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download region")

		return RegionResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from region was not 200")

		return RegionResponse{}, resp, errors.New("status was not 200")
	}

	region, err := NewRegionResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse region response")

		return RegionResponse{}, resp, err
	}

	return region, resp, nil
}

func NewRegionResponse(body []byte) (RegionResponse, error) {
	region := &RegionResponse{}
	if err := json.Unmarshal(body, region); err != nil {
		return RegionResponse{}, err
	}

	return *region, nil
}
