package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/inventorytype"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemquality"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const itemURLFormat = "https://%s/data/wow/item/%d?namespace=%s"

func DefaultGetItemURL(
	regionHostname string,
	id ItemId,
	version gameversion.GameVersion,
) (string, error) {
	namespace, err := gameversion.StaticNamespaceMap.Resolve(version)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(itemURLFormat, regionHostname, id, namespace), nil
}

type VersionItemTuple struct {
	GameVersion gameversion.GameVersion `json:"game_version"`
	Id          ItemId                  `json:"id"`
}

type ItemId int

type ItemQuality struct {
	Type itemquality.ItemQuality `json:"type"`
	Name locale.Mapping          `json:"name"`
}

type ItemMedia struct {
	Key HrefReference `json:"key"`
	Id  ItemId        `json:"id"`
}

type ItemInventoryType struct {
	Type inventorytype.InventoryType `json:"type"`
	Name locale.Mapping              `json:"name"`
}

type ItemSpellId int

type ItemResponse struct {
	LinksBase
	Id            ItemId            `json:"id"`
	Name          locale.Mapping    `json:"name"`
	Quality       ItemQuality       `json:"quality"`
	Level         int               `json:"level"`
	RequiredLevel int               `json:"required_level"`
	Media         ItemMedia         `json:"media"`
	ItemClass     ItemClass         `json:"item_class"`
	ItemSubClass  ItemSubClass      `json:"item_subclass"`
	InventoryType ItemInventoryType `json:"inventory_type"`
	PurchasePrice PriceValue        `json:"purchase_price"`
	SellPrice     PriceValue        `json:"sell_price"`
	MaxCount      int               `json:"max_count"`
	IsEquippable  bool              `json:"is_equippable"`
	IsStackable   bool              `json:"is_stackable"`

	// item-class-id: 9 (Recipe)
	Description locale.Mapping `json:"description"`

	PreviewItem ItemPreviewItem `json:"preview_item"`

	// unknown
	PurchaseQuantity int `json:"purchase_quantity"`
}

func NewItemFromHTTP(uri string) (ItemResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to download item")

		return ItemResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    uri,
		}).Error("resp from item was not 200")

		return ItemResponse{}, resp, errors.New("status was not 200")
	}

	item, err := NewItemResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to parse item response")

		return ItemResponse{}, resp, err
	}

	return item, resp, nil
}

func NewItemResponse(body []byte) (ItemResponse, error) {
	item := &ItemResponse{}
	if err := json.Unmarshal(body, item); err != nil {
		return ItemResponse{}, err
	}

	return *item, nil
}
