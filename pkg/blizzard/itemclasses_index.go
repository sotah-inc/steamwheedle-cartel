package blizzard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const itemClassIndexURLFormat = "https://%s/data/wow/item-class/index"

func DefaultGetItemClassIndexURL(regionHostname string) (string, error) {
	return fmt.Sprintf(itemClassIndexURLFormat, regionHostname), nil
}

type GetItemClassIndexURLFunc func(string) (string, error)

type ItemClassId int

type ItemClass struct {
	Key  HrefReference `json:"key"`
	Name string        `json:"string"`
	Id   ItemClassId   `json:"id"`
}

type ItemClassIndexResponse struct {
	Links struct {
		Self HrefReference `json:"self"`
	} `json:"_links"`
	ItemClasses []ItemClass `json:"item_classes"`
}

func NewItemClassIndexFromHTTP(uri string) (ItemClassIndexResponse, ResponseMeta, error) {
	resp, err := Download(uri)
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
