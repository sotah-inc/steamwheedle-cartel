package items

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

func NewQueryRequest(data []byte) (QueryRequest, error) {
	var out QueryRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return QueryRequest{}, err
	}

	return out, nil
}

type QueryRequest struct {
	Query  string        `json:"query"`
	Locale locale.Locale `json:"locale"`
}

type QueryItem struct {
	Target string            `json:"target"`
	ItemId blizzardv2.ItemId `json:"item_id"`
	Rank   int               `json:"rank"`
}

func NewQueryItems(
	idNormalizedNameMap sotah.ItemIdNameMap,
	providedLocale locale.Locale,
) (QueryItemList, error) {
	out := make(QueryItemList, len(idNormalizedNameMap))
	i := 0
	for id, normalizedName := range idNormalizedNameMap {
		foundName, ok := normalizedName[providedLocale]
		if !ok {
			return QueryItemList{}, fmt.Errorf("could not resolve normalized-name from locale %s", providedLocale)
		}

		out[i] = QueryItem{ItemId: id, Target: foundName}

		i += 1
	}

	return out, nil
}

type QueryItemList []QueryItem

func (iqItems QueryItemList) Limit() QueryItemList {
	listLength := len(iqItems)
	if listLength > 10 {
		listLength = 10
	}

	out := make(QueryItemList, listLength)
	for i := 0; i < listLength; i++ {
		out[i] = iqItems[i]
	}

	return out
}

func (iqItems QueryItemList) FilterLowRank() QueryItemList {
	out := QueryItemList{}
	for _, itemValue := range iqItems {
		if itemValue.Rank == -1 {
			continue
		}
		out = append(out, itemValue)
	}

	return out
}

type QueryItemListByTarget QueryItemList

func (by QueryItemListByTarget) Len() int           { return len(by) }
func (by QueryItemListByTarget) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryItemListByTarget) Less(i, j int) bool { return by[i].Target < by[j].Target }

type QueryItemListByRank QueryItemList

func (by QueryItemListByRank) Len() int           { return len(by) }
func (by QueryItemListByRank) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryItemListByRank) Less(i, j int) bool { return by[i].Rank < by[j].Rank }

type QueryResponse struct {
	Items QueryItemList `json:"items"`
}

func (r QueryResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (idBase Database) QueryItems(req QueryRequest) (QueryResponse, codes.Code, error) {
	// gathering items
	idNormalizedNameMap, err := idBase.GetIdNormalizedNameMap()
	if err != nil {
		return QueryResponse{}, codes.GenericError, err
	}

	// reformatting into query-items-items
	queryItems, err := NewQueryItems(idNormalizedNameMap, req.Locale)
	if err != nil {
		return QueryResponse{}, codes.UserError, err
	}

	// optionally sorting by rank or sorting by name
	if req.Query != "" {
		for i, iqItem := range queryItems {
			iqItem.Rank = fuzzy.RankMatchFold(req.Query, iqItem.Target)
			queryItems[i] = iqItem
		}
		queryItems = queryItems.FilterLowRank()
		sort.Sort(QueryItemListByRank(queryItems))
	} else {
		sort.Sort(QueryItemListByTarget(queryItems))
	}

	// truncating
	queryItems = queryItems.Limit()

	return QueryResponse{Items: queryItems}, codes.Ok, nil
}
