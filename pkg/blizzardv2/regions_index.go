package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const regionIndexURLFormat = "https://%s/data/wow/region/index?namespace=dynamic-%s"

func DefaultRegionIndexURL(regionHostname string, regionName RegionName) string {
	return fmt.Sprintf(regionIndexURLFormat, regionHostname, regionName)
}

type GetRegionIndexURLFunc func(string) string

type RegionIndexResponseRegion HrefReference

type RegionIndexResponse struct {
	LinksBase

	Regions []HrefReference `json:"regions"`
}

func NewRegionIndexFromHTTP(uri string) (RegionIndexResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download region-index")

		return RegionIndexResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from region-index was not 200")

		return RegionIndexResponse{}, resp, errors.New("status was not 200")
	}

	rIndex, err := NewRegionIndexResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse region-index response")

		return RegionIndexResponse{}, resp, err
	}

	return rIndex, resp, nil
}

func NewRegionIndexResponse(body []byte) (RegionIndexResponse, error) {
	rIndex := &RegionIndexResponse{}
	if err := json.Unmarshal(body, rIndex); err != nil {
		return RegionIndexResponse{}, err
	}

	return *rIndex, nil
}
