package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/binding"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/inventorytype"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemquality"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const itemURLFormat = "https://%s/data/wow/item/%d?namespace=static-%s"

func DefaultGetItemURL(regionHostname string, id ItemId, regionName RegionName) string {
	return fmt.Sprintf(itemURLFormat, regionHostname, id, regionName)
}

type GetItemURLFunc func(string, ItemId, RegionName) string

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

type ItemSpell struct {
	Spell struct {
		Key  HrefReference  `json:"key"`
		Name locale.Mapping `json:"name"`
		Id   ItemSpellId    `json:"id"`
	} `json:"spell"`
	Description locale.Mapping `json:"description"`
}

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
	PurchasePrice int64             `json:"purchase_price"`
	SellPrice     int64             `json:"sell_price"`
	MaxCount      int               `json:"max_count"`
	IsEquippable  bool              `json:"is_equippable"`
	IsStackable   bool              `json:"is_stackable"`
	PreviewItem   struct {
		Item struct {
			Key HrefReference `json:"key"`
			Id  ItemId        `json:"id"`
		} `json:"item"`
		Quality       ItemQuality       `json:"quality"`
		Name          locale.Mapping    `json:"name"`
		Media         ItemMedia         `json:"media"`
		ItemClass     ItemClass         `json:"item_class"`
		ItemSubClass  ItemSubClass      `json:"item_subclass"`
		InventoryType ItemInventoryType `json:"inventory_type"`
		Binding       struct {
			Type binding.Binding `json:"type"`
			Name locale.Mapping  `json:"name"`
		} `json:"binding"`
		Spells    []ItemSpell `json:"spells"`
		SellPrice struct {
			Value  int64          `json:"value"`
			Header locale.Mapping `json:"header"`
			Gold   locale.Mapping `json:"gold"`
			Silver locale.Mapping `json:"silver"`
			Copper locale.Mapping `json:"copper"`
		} `json:"sell_price"`
		IsSubClassHidden bool `json:"is_subclass_hidden"`
		NameDescription  struct {
			DisplayString locale.Mapping `json:"display_string"`
			Color         struct {
				Red   int     `json:"r"`
				Green int     `json:"g"`
				Blue  int     `json:"b"`
				Alpha float32 `json:"a"`
			} `json:"color"`
		} `json:"name_description"`
	} `json:"preview_item"`
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
