package blizzardv2

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

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

func NewItemMediaResponse(body []byte) (ItemMediaResponse, error) {
	iMedia := &ItemMediaResponse{}
	if err := json.Unmarshal(body, iMedia); err != nil {
		return ItemMediaResponse{}, err
	}

	return *iMedia, nil
}

type ItemMediaAsset struct {
	Key string `json:"key"`
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

	k := res.Assets[0].Key

	lastSlashIndex := strings.LastIndex(k, "/")
	if lastSlashIndex == -1 {
		return "", errors.New("asset key did not have slash")
	}

	lastDotIndex := strings.LastIndex(k, ".")
	if lastDotIndex == -1 {
		return "", errors.New("asset key did not have dot")
	}

	return k[lastSlashIndex+1 : lastDotIndex], nil
}
