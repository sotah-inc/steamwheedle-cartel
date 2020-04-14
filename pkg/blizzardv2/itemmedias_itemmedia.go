package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const itemMediaURLFormat = "https://%s/data/wow/media/item/%d?namespace=static-%s"

func DefaultGetItemMediaURL(regionHostname string, id ItemId, regionName RegionName) string {
	return fmt.Sprintf(itemMediaURLFormat, regionHostname, id, regionName)
}

type GetItemMediaURLFunc func(string, ItemId, RegionName) string

type ItemMediaAsset struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ItemMediaResponse struct {
	LinksBase
	Assets []ItemMediaAsset `json:"assets"`
	Id     ItemId           `json:"id"`
}

func (res ItemMediaResponse) GetIcon() (string, error) {
	if len(res.Assets) == 0 {
		return "", errors.New("could not find ")
	}

	v := res.Assets[0].Value

	lastSlashIndex := strings.LastIndex(v, "/")
	if lastSlashIndex == -1 {
		return "", errors.New("asset key did not have slash")
	}

	lastDotIndex := strings.LastIndex(v, ".")
	if lastDotIndex == -1 {
		return "", errors.New("asset key did not have dot")
	}

	return v[lastSlashIndex+1 : lastDotIndex], nil
}

func NewItemMediaResponse(body []byte) (ItemMediaResponse, error) {
	iMedia := &ItemMediaResponse{}
	if err := json.Unmarshal(body, iMedia); err != nil {
		return ItemMediaResponse{}, err
	}

	return *iMedia, nil
}

func NewItemMediaFromHTTP(uri string) (ItemMediaResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to download item-media")

		return ItemMediaResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    uri,
		}).Error("resp from item-media was not 200")

		return ItemMediaResponse{}, resp, errors.New("status was not 200")
	}

	item, err := NewItemMediaResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to parse item-media response")

		return ItemMediaResponse{}, resp, err
	}

	return item, resp, nil
}
