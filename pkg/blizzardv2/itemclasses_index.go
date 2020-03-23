package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const itemClassIndexURLFormat = "https://%s/data/wow/item-class/index?namespace=static-us"

func DefaultGetItemClassIndexURL(regionHostname string) (string, error) {
	return fmt.Sprintf(itemClassIndexURLFormat, regionHostname), nil
}

type GetItemClassIndexURLFunc func(string) (string, error)

type ItemClassId int

type ItemClass struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"string"`
	Id   ItemClassId    `json:"id"`
}

type ItemClassIndexResponse struct {
	LinksBase
	ItemClasses []ItemClass `json:"item_classes"`
}

func NewItemClassIndexFromHTTP(uri string) (ItemClassIndexResponse, blizzard.ResponseMeta, error) {
	resp, err := blizzard.Download(uri)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to download item-class-index")

		return ItemClassIndexResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    uri,
		}).Error("resp from item-class-index was not 200")

		return ItemClassIndexResponse{}, resp, errors.New("status was not 200")
	}

	icIndex, err := NewItemClassIndexResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
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
