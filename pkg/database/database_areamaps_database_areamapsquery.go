package database

import (
	"encoding/json"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func NewAreaMapsQueryRequest(data []byte) (AreaMapsQueryRequest, error) {
	var out AreaMapsQueryRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return AreaMapsQueryRequest{}, err
	}

	return out, nil
}

type AreaMapsQueryRequest struct {
	Query string `json:"query"`
}

type AreaMapsQueryResult struct {
	Target    string          `json:"target"`
	AreaMapId sotah.AreaMapId `json:"areamap_id"`
	Rank      int             `json:"rank"`
}

func NewAreaMapsQueryResultList(idNormalizedNameMap sotah.AreaMapIdNameMap) AreaMapsQueryResultList {
	out := AreaMapsQueryResultList{}
	for id, normalizedName := range idNormalizedNameMap {
		if normalizedName == "" {
			continue
		}

		out = append(out, AreaMapsQueryResult{
			AreaMapId: id,
			Target:    normalizedName,
		})
	}

	return out
}

type AreaMapsQueryResultList []AreaMapsQueryResult

func (iqAreaMaps AreaMapsQueryResultList) Limit() AreaMapsQueryResultList {
	listLength := len(iqAreaMaps)
	if listLength > 10 {
		listLength = 10
	}

	out := make(AreaMapsQueryResultList, listLength)
	for i := 0; i < listLength; i++ {
		out[i] = iqAreaMaps[i]
	}

	return out
}

func (iqAreaMaps AreaMapsQueryResultList) FilterLowRank() AreaMapsQueryResultList {
	out := AreaMapsQueryResultList{}
	for _, itemValue := range iqAreaMaps {
		if itemValue.Rank == -1 {
			continue
		}

		out = append(out, itemValue)
	}

	return out
}

type AreaMapsQueryResultListByTarget AreaMapsQueryResultList

func (by AreaMapsQueryResultListByTarget) Len() int           { return len(by) }
func (by AreaMapsQueryResultListByTarget) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by AreaMapsQueryResultListByTarget) Less(i, j int) bool { return by[i].Target < by[j].Target }

type AreaMapsQueryResultListByRank AreaMapsQueryResultList

func (by AreaMapsQueryResultListByRank) Len() int           { return len(by) }
func (by AreaMapsQueryResultListByRank) Swap(i, j int)      { by[i], by[j] = by[j], by[i] }
func (by AreaMapsQueryResultListByRank) Less(i, j int) bool { return by[i].Rank < by[j].Rank }

type AreaMapsQueryResponse struct {
	Results AreaMapsQueryResultList `json:"result"`
}

func (r AreaMapsQueryResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(r)
}

func (r AreaMapsQueryResponse) Ids() []sotah.AreaMapId {
	out := make([]sotah.AreaMapId, len(r.Results))
	i := 0
	for _, result := range r.Results {
		out[i] = result.AreaMapId

		i += 1
	}

	return out
}

func (amBase AreaMapsDatabase) AreaMapsQuery(req AreaMapsQueryRequest) (AreaMapsQueryResponse, error) {
	// gathering area-maps
	idNormalizedNameMap, err := amBase.GetIdNormalizedNameMap()
	if err != nil {
		return AreaMapsQueryResponse{}, err
	}

	// reformatting into area-maps query result list
	results := NewAreaMapsQueryResultList(idNormalizedNameMap)

	// optionally sorting by rank or sorting by name
	if req.Query != "" {
		for i, iqAreaMap := range results {
			iqAreaMap.Rank = fuzzy.RankMatchFold(req.Query, iqAreaMap.Target)
			results[i] = iqAreaMap
		}
		results = results.FilterLowRank()
		sort.Sort(AreaMapsQueryResultListByRank(results))
	} else {
		sort.Sort(AreaMapsQueryResultListByTarget(results))
	}

	// truncating
	results = results.Limit()

	return AreaMapsQueryResponse{results}, nil
}
