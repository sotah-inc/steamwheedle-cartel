package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const itemClassIndexURLFormat = "https://%s/data/wow/item-class/index?namespace=static-%s"

func DefaultGetItemClassIndexURL(regionHostname string, regionName RegionName) string {
	return fmt.Sprintf(itemClassIndexURLFormat, regionHostname, regionName)
}

type GetItemClassIndexURLFunc func(string, RegionName) string

type ItemClass struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"string"`
	Id   itemclass.Id   `json:"id"`
}

type ItemClassIndexResponse struct {
	LinksBase
	ItemClasses []ItemClass `json:"item_classes"`
}

func NewItemClassIndexFromHTTP(uri string) (ItemClassIndexResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download item-class-index")

		return ItemClassIndexResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from item-class-index was not 200")

		return ItemClassIndexResponse{}, resp, errors.New("status was not 200")
	}

	icIndex, err := NewItemClassIndexResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse item-class-index response")

		return ItemClassIndexResponse{}, resp, err
	}

	return icIndex, resp, nil
}

func NewItemClassIndexResponse(body []byte) (ItemClassIndexResponse, error) {
	icIndex := &ItemClassIndexResponse{}
	if err := json.Unmarshal(body, icIndex); err != nil {
		return ItemClassIndexResponse{}, err
	}

	return *icIndex, nil
}
