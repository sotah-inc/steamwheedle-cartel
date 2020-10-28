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

func NewQueryPetsRequest(data []byte) (QueryPetsRequest, error) {
	var out QueryPetsRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return QueryPetsRequest{}, err
	}

	return out, nil
}

type QueryPetsRequest struct {
	Query  string        `json:"query"`
	Locale locale.Locale `json:"locale"`
}

type QueryPetsItem struct {
	Target string           `json:"target"`
	PetId  blizzardv2.PetId `json:"pet_id"`
	Rank   int              `json:"rank"`
}

func NewQueryPetsItems(
	idNormalizedNameMap sotah.PetIdNameMap,
	providedLocale locale.Locale,
) (QueryPetsItems, error) {
	out := make(QueryPetsItems, len(idNormalizedNameMap))
	i := 0
	for id, normalizedName := range idNormalizedNameMap {
		foundName, ok := normalizedName[providedLocale]
		if !ok {
			return QueryPetsItems{}, fmt.Errorf("could not resolve normalized-name from locale %s", providedLocale)
		}

		out[i] = QueryPetsItem{PetId: id, Target: foundName}

		i += 1
	}

	return out, nil
}

type QueryPetsItems []QueryPetsItem

func (qpItems QueryPetsItems) Limit() QueryPetsItems {
	listLength := len(qpItems)
	if listLength > 10 {
		listLength = 10
	}

	out := make(QueryPetsItems, listLength)
	for i := 0; i < listLength; i++ {
		out[i] = qpItems[i]
	}

	return out
}

func (qpItems QueryPetsItems) FilterLowRank() QueryPetsItems {
	out := QueryPetsItems{}
	for _, itemValue := range qpItems {
		if itemValue.Rank == -1 {
			continue
		}

		out = append(out, itemValue)
	}

	return out
}

type QueryPetsItemsByTarget QueryPetsItems

func (by QueryPetsItemsByTarget) Len() int           { return len(by) }
func (by QueryPetsItemsByTarget) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryPetsItemsByTarget) Less(i, j int) bool { return by[i].Target < by[j].Target }

type QueryPetsItemsByRank QueryPetsItems

func (by QueryPetsItemsByRank) Len() int           { return len(by) }
func (by QueryPetsItemsByRank) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryPetsItemsByRank) Less(i, j int) bool { return by[i].Rank < by[j].Rank }

type QueryPetsResponse struct {
	Items QueryPetsItems `json:"items"`
}

func (r QueryPetsResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (pdBase PetsDatabase) QueryPets(req QueryPetsRequest) (QueryPetsResponse, codes.Code, error) {
	// gathering pets
	idNormalizedNameMap, err := pdBase.GetIdNormalizedNameMap()
	if err != nil {
		return QueryPetsResponse{}, codes.GenericError, err
	}

	// reformatting into query-pets-items
	QueryPets, err := NewQueryPetsItems(idNormalizedNameMap, req.Locale)
	if err != nil {
		return QueryPetsResponse{}, codes.UserError, err
	}

	// optionally sorting by rank or sorting by name
	if req.Query != "" {
		for i, qpItem := range QueryPets {
			qpItem.Rank = fuzzy.RankMatchFold(req.Query, qpItem.Target)
			QueryPets[i] = qpItem
		}
		QueryPets = QueryPets.FilterLowRank()
		sort.Sort(QueryPetsItemsByRank(QueryPets))
	} else {
		sort.Sort(QueryPetsItemsByTarget(QueryPets))
	}

	// truncating
	QueryPets = QueryPets.Limit()

	return QueryPetsResponse{Items: QueryPets}, codes.Ok, nil
}
