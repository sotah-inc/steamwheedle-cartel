package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const itemClassURLFormat = "https://%s/data/wow/item-class/%d?namespace=static-%s"

func DefaultGetItemClassURL(regionHostname string, regionName RegionName, id itemclass.Id) string {
	return fmt.Sprintf(itemClassURLFormat, regionHostname, id, regionName)
}

type GetItemClassURLFunc func(string, RegionName, itemclass.Id) string

type ItemSubClassId int

type ItemSubClass struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   ItemSubClassId `json:"id"`
}

type ItemClassResponse struct {
	LinksBase
	ClassId        itemclass.Id   `json:"class_id"`
	Name           locale.Mapping `json:"name"`
	ItemSubClasses []ItemSubClass `json:"item_subclasses"`
}

func NewItemClassResponse(body []byte) (ItemClassResponse, error) {
	iClass := &ItemClassResponse{}
	if err := json.Unmarshal(body, iClass); err != nil {
		return ItemClassResponse{}, err
	}

	return *iClass, nil
}

func NewItemClassFromHTTP(uri string) (ItemClassResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download item-class")

		return ItemClassResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status":  resp.Status,
			"uri":     ClearAccessToken(uri),
			"body":    string(resp.Body),
			"headers": resp.Headers,
		}).Error("resp from item-class was not 200")

		return ItemClassResponse{}, resp, errors.New("status was not 200")
	}

	iClass, err := NewItemClassResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse item-class response")

		return ItemClassResponse{}, resp, err
	}

	return iClass, resp, nil
}
