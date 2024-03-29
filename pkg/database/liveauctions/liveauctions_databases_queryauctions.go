package liveauctions

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortdirections"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/sortkinds"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewQueryRequest(data []byte) (QueryRequest, error) {
	ar := &QueryRequest{}
	err := json.Unmarshal(data, &ar)
	if err != nil {
		return QueryRequest{}, err
	}

	return *ar, nil
}

type QueryRequest struct {
	Tuple         blizzardv2.RegionVersionConnectedRealmTuple `json:"tuple"`
	Page          int                                         `json:"page"`
	Count         int                                         `json:"count"`
	SortDirection sortdirections.SortDirection                `json:"sort_direction"`
	SortKind      sortkinds.SortKind                          `json:"sort_kind"`
	ItemFilters   blizzardv2.ItemIds                          `json:"item_filters"`
	PetFilters    []blizzardv2.PetId                          `json:"pet_filters"`
}

type QueryResponse struct {
	AuctionList sotah.MiniAuctionList `json:"auctions"`
	Total       int                   `json:"total"`
	TotalCount  int                   `json:"total_count"`
}

func (qr QueryResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(qr)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (ladBases Databases) QueryAuctions(
	qr QueryRequest,
) (QueryResponse, codes.Code, error) {
	ladBase, err := ladBases.GetDatabase(qr.Tuple)
	if err != nil {
		return QueryResponse{}, codes.UserError, err
	}

	if qr.Page < 0 {
		return QueryResponse{}, codes.UserError, errors.New("page must be >= 0")
	}
	if qr.Count == 0 {
		return QueryResponse{}, codes.UserError, errors.New("count must be >= 0")
	} else if qr.Count > 1000 {
		return QueryResponse{}, codes.UserError, errors.New("page must be <= 1000")
	}

	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return QueryResponse{}, codes.GenericError, err
	}

	// initial response format
	aResponse := QueryResponse{Total: -1, TotalCount: -1, AuctionList: maList}

	filterCriteria := sotah.MiniAuctionListFilterCriteria{
		ItemIds: qr.ItemFilters,
		PetIds:  qr.PetFilters,
	}
	if !filterCriteria.IsEmpty() {
		mafList := sotah.NewMiniAuctionFlaggedList(maList)
		aResponse.AuctionList = mafList.Flag(filterCriteria).FilterInFlagged().ToMiniAuctionList()
	}

	// calculating the total for paging
	aResponse.Total = len(aResponse.AuctionList)

	// calculating the total-count for review
	totalCount := 0
	for _, mAuction := range maList {
		totalCount += len(mAuction.AucList)
	}
	aResponse.TotalCount = totalCount

	// optionally sorting
	if qr.SortKind != sortkinds.None && qr.SortDirection != sortdirections.None {
		err = aResponse.AuctionList.Sort(qr.SortKind, qr.SortDirection)
		if err != nil {
			return QueryResponse{}, codes.UserError, err
		}
	}

	// truncating the list
	aResponse.AuctionList, err = aResponse.AuctionList.Limit(qr.Count, qr.Page)
	if err != nil {
		return QueryResponse{}, codes.UserError, err
	}

	return aResponse, codes.Ok, nil
}
