package areamaps

import (
	"encoding/json"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
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
	Query string `json:"query"`
}

type QueryResult struct {
	Target    string          `json:"target"`
	AreaMapId sotah.AreaMapId `json:"areamap_id"`
	Rank      int             `json:"rank"`
}

func NewQueryResultList(idNormalizedNameMap sotah.AreaMapIdNameMap) QueryResultList {
	out := QueryResultList{}
	for id, normalizedName := range idNormalizedNameMap {
		if normalizedName == "" {
			continue
		}

		out = append(out, QueryResult{
			AreaMapId: id,
			Target:    normalizedName,
		})
	}

	return out
}

type QueryResultList []QueryResult

func (iqAreaMaps QueryResultList) Limit() QueryResultList {
	listLength := len(iqAreaMaps)
	if listLength > 10 {
		listLength = 10
	}

	out := make(QueryResultList, listLength)
	for i := 0; i < listLength; i++ {
		out[i] = iqAreaMaps[i]
	}

	return out
}

func (iqAreaMaps QueryResultList) FilterLowRank() QueryResultList {
	out := QueryResultList{}
	for _, itemValue := range iqAreaMaps {
		if itemValue.Rank == -1 {
			continue
		}

		out = append(out, itemValue)
	}

	return out
}

type QueryResultListByTarget QueryResultList

func (by QueryResultListByTarget) Len() int           { return len(by) }
func (by QueryResultListByTarget) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryResultListByTarget) Less(i, j int) bool { return by[i].Target < by[j].Target }

type QueryResultListByRank QueryResultList

func (by QueryResultListByRank) Len() int           { return len(by) }
func (by QueryResultListByRank) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by QueryResultListByRank) Less(i, j int) bool { return by[i].Rank < by[j].Rank }

type QueryResponse struct {
	Results QueryResultList `json:"result"`
}

func (r QueryResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (r QueryResponse) Ids() []sotah.AreaMapId {
	out := make([]sotah.AreaMapId, len(r.Results))
	i := 0
	for _, result := range r.Results {
		out[i] = result.AreaMapId

		i += 1
	}

	return out
}

func (amBase Database) AreaMapsQuery(req QueryRequest) (QueryResponse, error) {
	// gathering area-maps
	idNormalizedNameMap, err := amBase.GetIdNormalizedNameMap()
	if err != nil {
		return QueryResponse{}, err
	}

	// reformatting into area-maps query result list
	results := NewQueryResultList(idNormalizedNameMap)

	// optionally sorting by rank or sorting by name
	if req.Query != "" {
		for i, iqAreaMap := range results {
			iqAreaMap.Rank = fuzzy.RankMatchFold(req.Query, iqAreaMap.Target)
			results[i] = iqAreaMap
		}
		results = results.FilterLowRank()
		sort.Sort(QueryResultListByRank(results))
	} else {
		sort.Sort(QueryResultListByTarget(results))
	}

	// truncating
	results = results.Limit()

	return QueryResponse{results}, nil
}
