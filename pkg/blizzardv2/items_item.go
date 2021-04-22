package blizzardv2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	"github.com/sirupsen/logrus"
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

func NewItemClassItemsMap(ids []itemclass.Id) ItemClassItemsMap {
	out := ItemClassItemsMap{}
	for _, id := range ids {
		out[id] = ItemIds{}
	}

	return out
}

type ItemClassItemsMap map[itemclass.Id]ItemIds

func (iciMap ItemClassItemsMap) Find(classId itemclass.Id) ItemIds {
	found, ok := iciMap[classId]
	if !ok {
		return ItemIds{}
	}

	return found
}

func (iciMap ItemClassItemsMap) Insert(
	providedClassId itemclass.Id,
	providedItemId ItemId,
) ItemClassItemsMap {
	iciMap[providedClassId] = iciMap.Find(providedClassId).Merge(ItemIds{providedItemId})

	return iciMap
}

func (iciMap ItemClassItemsMap) ItemClassIds() []itemclass.Id {
	out := make([]itemclass.Id, len(iciMap))
	i := 0
	for id := range iciMap {
		out[i] = id

		i += 1
	}

	return out
}

type ItemId int

func NewItemRecipesMap(base64Encoded string) (ItemRecipesMap, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return ItemRecipesMap{}, err
	}

	jsonEncoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return ItemRecipesMap{}, err
	}

	out := ItemRecipesMap{}
	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return ItemRecipesMap{}, err
	}

	return out, nil
}

type ItemRecipesMap map[ItemId]RecipeIds

func (irMap ItemRecipesMap) ItemIds() []ItemId {
	out := make([]ItemId, len(irMap))
	i := 0
	for id := range irMap {
		out[i] = id

		i += 1
	}

	return out
}

func (irMap ItemRecipesMap) Find(id ItemId) RecipeIds {
	found, ok := irMap[id]
	if !ok {
		return RecipeIds{}
	}

	return found
}

func (irMap ItemRecipesMap) Merge(input ItemRecipesMap) ItemRecipesMap {
	out := ItemRecipesMap{}
	for id, recipeIds := range irMap {
		out[id] = recipeIds
	}
	for id, providedRecipeIds := range input {
		foundRecipeIds := irMap.Find(id)
		out[id] = foundRecipeIds.Merge(providedRecipeIds)
	}

	return out
}

func (irMap ItemRecipesMap) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(irMap)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (irMap ItemRecipesMap) FilterBlank() ItemRecipesMap {
	out := ItemRecipesMap{}
	for itemId, recipeIds := range irMap {
		if len(recipeIds) == 0 {
			continue
		}

		out[itemId] = recipeIds
	}

	return out
}

func (irMap ItemRecipesMap) ToRecipesItemMap() map[RecipeId]ItemId {
	out := map[RecipeId]ItemId{}
	for itemId, recipeIds := range irMap {
		for _, recipeId := range recipeIds {
			out[recipeId] = itemId
		}
	}

	return out
}

func (irMap ItemRecipesMap) RecipeIds() RecipeIds {
	outMap := map[RecipeId]struct{}{}
	for _, recipeIds := range irMap {
		for _, recipeId := range recipeIds {
			outMap[recipeId] = struct{}{}
		}
	}

	out := make(RecipeIds, len(outMap))
	i := 0
	for recipeId := range outMap {
		out[i] = recipeId

		i += 1
	}

	return out
}

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
