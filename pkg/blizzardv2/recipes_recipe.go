package blizzardv2

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const recipeURLFormat = "https://%s/data/wow/recipe/%d?namespace=static-%s"

func DefaultGetRecipeURL(regionHostname string, id RecipeId, regionName RegionName) string {
	return fmt.Sprintf(recipeURLFormat, regionHostname, id, regionName)
}

type GetRecipeURLFunc func(string, RecipeId, RegionName) string

func NewRecipeIdDescriptionMap(base64Encoded string) (RecipeIdDescriptionMap, error) {
	gzipEncoded, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		return RecipeIdDescriptionMap{}, err
	}

	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return RecipeIdDescriptionMap{}, err
	}

	out := RecipeIdDescriptionMap{}
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return RecipeIdDescriptionMap{}, err
	}

	return out, nil
}

type RecipeIdDescriptionMap map[RecipeId]string

func (rdMap RecipeIdDescriptionMap) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(rdMap)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func NewRecipeIdsMap(input RecipeIds) RecipeIdsMap {
	inputMap := map[RecipeId]struct{}{}
	for _, id := range input {
		inputMap[id] = struct{}{}
	}

	return inputMap
}

type RecipeIdsMap map[RecipeId]struct{}

func (idsMap RecipeIdsMap) ToIds() RecipeIds {
	out := make([]RecipeId, len(idsMap))
	i := 0
	for id := range idsMap {
		out[i] = id

		i += 1
	}

	return out
}

func NewRecipeIdsFromJson(jsonEncoded []byte) (RecipeIds, error) {
	out := RecipeIds{}

	if err := json.Unmarshal(jsonEncoded, &out); err != nil {
		return RecipeIds{}, err
	}

	return out, nil
}

type RecipeIds []RecipeId

func (ids RecipeIds) Append(input RecipeIds) RecipeIds {
	out := make(RecipeIds, len(ids)+len(input))
	i := 0
	for _, id := range ids {
		out[i] = id

		i += 1
	}

	for _, id := range input {
		out[i] = id

		i += 1
	}

	return out
}

func (ids RecipeIds) IsZero() bool {
	return len(ids) == 0
}

func (ids RecipeIds) Subtract(input RecipeIds) RecipeIds {
	inputMap := NewRecipeIdsMap(input)

	out := RecipeIds{}
	for _, id := range ids {
		if _, ok := inputMap[id]; ok {
			continue
		}

		out = append(out, id)
	}

	return out
}

func (ids RecipeIds) Merge(input RecipeIds) RecipeIds {
	inputMap := NewRecipeIdsMap(input)
	for _, id := range input {
		inputMap[id] = struct{}{}
	}

	return inputMap.ToIds()
}

func (ids RecipeIds) EncodeForStorage() ([]byte, error) {
	return json.Marshal(ids)
}

type RecipeId int

type RecipeItem struct {
	Key  HrefReference  `json:"key"`
	Name locale.Mapping `json:"name"`
	Id   ItemId         `json:"id"`
}

func (item RecipeItem) IsZero() bool {
	return item.Id == 0
}

type RecipeReagent struct {
	Reagent  RecipeItem `json:"reagent"`
	Quantity int        `json:"quantity"`
}

type RecipeModifiedCraftingSlots struct {
	SlotType struct {
		Key HrefReference `json:"key"`
		Id  int           `json:"id"`
	} `json:"slot_type"`
	DisplayOrder int `json:"display_order"`
}

type RecipeResponse struct {
	LinksBase
	Id          RecipeId       `json:"id"`
	Name        locale.Mapping `json:"name"`
	Description locale.Mapping `json:"description"`
	Media       struct {
		Key HrefReference `json:"key"`
		Id  RecipeId      `json:"id"`
	} `json:"media"`
	CraftedItem         RecipeItem      `json:"crafted_item"`
	AllianceCraftedItem RecipeItem      `json:"alliance_crafted_item"`
	HordeCraftedItem    RecipeItem      `json:"horde_crafted_item"`
	Reagents            []RecipeReagent `json:"reagents"`
	Rank                int             `json:"rank"`
	CraftedQuantity     struct {
		Value float32 `json:"value"`
	} `json:"crafted_quantity"`
	ModifiedCraftingSlots []RecipeModifiedCraftingSlots `json:"modified_crafting_slots"`
}

func NewRecipeResponse(body []byte) (RecipeResponse, error) {
	psTier := &RecipeResponse{}
	if err := json.Unmarshal(body, psTier); err != nil {
		return RecipeResponse{}, err
	}

	return *psTier, nil
}

func NewRecipeResponseFromHTTP(uri string) (RecipeResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download recipe")

		return RecipeResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from recipe was not 200")

		return RecipeResponse{}, resp, errors.New("status was not 200")
	}

	psTier, err := NewRecipeResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse recipe response")

		return RecipeResponse{}, resp, err
	}

	return psTier, resp, nil
}
