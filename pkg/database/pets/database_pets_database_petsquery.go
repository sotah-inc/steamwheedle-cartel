package pets

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
	Target string           `json:"target"`
	PetId  blizzardv2.PetId `json:"pet_id"`
	Rank   int              `json:"rank"`
}

func NewQueryItems(
	idNormalizedNameMap sotah.PetIdNameMap,
	providedLocale locale.Locale,
) (QueryItems, error) {
	out := make(QueryItems, len(idNormalizedNameMap))
	i := 0
	for id, normalizedName := range idNormalizedNameMap {
		foundName, ok := normalizedName[providedLocale]
		if !ok {
			return QueryItems{}, fmt.Errorf("could not resolve normalized-name from locale %s", providedLocale)
		}

		out[i] = QueryItem{PetId: id, Target: foundName}

		i += 1
	}

	return out, nil
}

type QueryItems []QueryItem

func (qpItems QueryItems) Limit() QueryItems {
	listLength := len(qpItems)
	if listLength > 10 {
		listLength = 10
	}

	out := make(QueryItems, listLength)
	for i := 0; i < listLength; i++ {
		out[i] = qpItems[i]
	}

	return out
}

func (qpItems QueryItems) FilterLowRank() QueryItems {
	out := QueryItems{}
	for _, itemValue := range qpItems {
		if itemValue.Rank == -1 {
			continue
		}

		out = append(out, itemValue)
	}

	return out
}

type QueryItemsByTarget QueryItems

func (by QueryItemsByTarget) Len() int           { return len(by) }
func (by QueryItemsByTarget) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryItemsByTarget) Less(i, j int) bool { return by[i].Target < by[j].Target }

type QueryItemsByRank QueryItems

func (by QueryItemsByRank) Len() int           { return len(by) }
func (by QueryItemsByRank) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryItemsByRank) Less(i, j int) bool { return by[i].Rank < by[j].Rank }

type QueryResponse struct {
	Items QueryItems `json:"items"`
}

func (r QueryResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (pdBase Database) QueryPets(req QueryRequest) (QueryResponse, codes.Code, error) {
	// gathering pets
	idNormalizedNameMap, err := pdBase.GetIdNormalizedNameMap()
	if err != nil {
		return QueryResponse{}, codes.GenericError, err
	}

	// reformatting into query-pets-items
	QueryPets, err := NewQueryItems(idNormalizedNameMap, req.Locale)
	if err != nil {
		return QueryResponse{}, codes.UserError, err
	}

	// optionally sorting by rank or sorting by name
	if req.Query != "" {
		for i, qpItem := range QueryPets {
			qpItem.Rank = fuzzy.RankMatchFold(req.Query, qpItem.Target)
			QueryPets[i] = qpItem
		}
		QueryPets = QueryPets.FilterLowRank()
		sort.Sort(QueryItemsByRank(QueryPets))
	} else {
		sort.Sort(QueryItemsByTarget(QueryPets))
	}

	// truncating
	QueryPets = QueryPets.Limit()

	return QueryResponse{Items: QueryPets}, codes.Ok, nil
}
