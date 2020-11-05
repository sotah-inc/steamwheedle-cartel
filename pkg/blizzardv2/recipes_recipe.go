package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const recipeURLFormat = "https://%s/data/wow/recipe/%d?namespace=static-%s"

func DefaultGetRecipeURL(regionHostname string, id RecipeId, regionName RegionName) string {
	return fmt.Sprintf(recipeURLFormat, regionHostname, id, regionName)
}

type GetRecipeURLFunc func(string, RecipeId, RegionName) string

func NewRecipeIdsFromMap(idMap RecipeIdMap) RecipeIds {
	out := make(RecipeIds, len(idMap))
	i := 0
	for id := range idMap {
		out[i] = id

		i += 1
	}

	return out
}

type RecipeIds []RecipeId

func (ids RecipeIds) Merge(input RecipeIds) RecipeIds {
	idMap := NewRecipeIdMap(ids)
	for _, id := range input {
		idMap[id] = struct{}{}
	}

	return NewRecipeIdsFromMap(idMap)
}

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

func NewRecipeIdMap(ids RecipeIds) RecipeIdMap {
	out := RecipeIdMap{}
	for _, id := range ids {
		out[id] = struct{}{}
	}

	return out
}

type RecipeIdMap map[RecipeId]struct{}

func (idMap RecipeIdMap) Exclude(ids RecipeIds) RecipeIdMap {
	excludedIdMap := NewRecipeIdMap(ids)
	out := RecipeIdMap{}
	for id := range idMap {
		if _, ok := excludedIdMap[id]; ok {
			continue
		}

		out[id] = struct{}{}
	}

	return out
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
	AllianceCraftedItem RecipeItem      `json:"alliance_crafted_item"`
	HordeCraftedItem    RecipeItem      `json:"horde_crafted_item"`
	Reagents            []RecipeReagent `json:"reagents"`
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
