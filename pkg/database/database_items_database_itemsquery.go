package database

import (
	"encoding/json"
	"fmt"
	"sort"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewQueryItemsRequest(data []byte) (QueryItemsRequest, error) {
	var out QueryItemsRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return QueryItemsRequest{}, err
	}

	return out, nil
}

type QueryItemsRequest struct {
	Query  string        `json:"query"`
	Locale locale.Locale `json:"locale"`
}
type QueryItemsItem struct {
	Target string            `json:"target"`
	ItemId blizzardv2.ItemId `json:"item_id"`
	Rank   int               `json:"rank"`
}

func NewQueryItemsItems(
	idNormalizedNameMap sotah.ItemIdNameMap,
	providedLocale locale.Locale,
) (QueryItemsItems, error) {
	out := make(QueryItemsItems, len(idNormalizedNameMap))
	i := 0
	for id, normalizedName := range idNormalizedNameMap {
		foundName, ok := normalizedName[providedLocale]
		if !ok {
			return QueryItemsItems{}, fmt.Errorf("could not resolve normalized-name from locale %s", providedLocale)
		}

		out[i] = QueryItemsItem{ItemId: id, Target: foundName}

		i += 1
	}

	return out, nil
}

type QueryItemsItems []QueryItemsItem

func (iqItems QueryItemsItems) Limit() QueryItemsItems {
	listLength := len(iqItems)
	if listLength > 10 {
		listLength = 10
	}

	out := make(QueryItemsItems, listLength)
	for i := 0; i < listLength; i++ {
		out[i] = iqItems[i]
	}

	return out
}

func (iqItems QueryItemsItems) FilterLowRank() QueryItemsItems {
	out := QueryItemsItems{}
	for _, itemValue := range iqItems {
		if itemValue.Rank == -1 {
			continue
		}
		out = append(out, itemValue)
	}

	return out
}

type QueryItemsItemsByTarget QueryItemsItems

func (by QueryItemsItemsByTarget) Len() int           { return len(by) }
func (by QueryItemsItemsByTarget) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryItemsItemsByTarget) Less(i, j int) bool { return by[i].Target < by[j].Target }

type QueryItemsItemsByRank QueryItemsItems

func (by QueryItemsItemsByRank) Len() int           { return len(by) }
func (by QueryItemsItemsByRank) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryItemsItemsByRank) Less(i, j int) bool { return by[i].Rank < by[j].Rank }

type QueryItemsResponse struct {
	Items QueryItemsItems `json:"items"`
}

func (r QueryItemsResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (idBase ItemsDatabase) QueryItems(req QueryItemsRequest) (QueryItemsResponse, codes.Code, error) {
	// gathering items
	idNormalizedNameMap, err := idBase.GetIdNormalizedNameMap()
	if err != nil {
		return QueryItemsResponse{}, codes.GenericError, err
	}

	// reformatting into query-items-items
	queryItems, err := NewQueryItemsItems(idNormalizedNameMap, req.Locale)
	if err != nil {
		return QueryItemsResponse{}, codes.UserError, err
	}

	// optionally sorting by rank or sorting by name
	if req.Query != "" {
		for i, iqItem := range queryItems {
			iqItem.Rank = fuzzy.RankMatchFold(req.Query, iqItem.Target)
			queryItems[i] = iqItem
		}
		queryItems = queryItems.FilterLowRank()
		sort.Sort(QueryItemsItemsByRank(queryItems))
	} else {
		sort.Sort(QueryItemsItemsByTarget(queryItems))
	}

	// truncating
	queryItems = queryItems.Limit()

	return QueryItemsResponse{Items: queryItems}, codes.Ok, nil
}
