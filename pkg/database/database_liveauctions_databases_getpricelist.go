package database

import (
	"encoding/base64"
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewGetPricelistRequest(data []byte) (GetPricelistRequest, error) {
	plRequest := &GetPricelistRequest{}
	err := json.Unmarshal(data, &plRequest)
	if err != nil {
		return GetPricelistRequest{}, err
	}

	return *plRequest, nil
}

type GetPricelistRequest struct {
	Tuple   blizzardv2.RegionConnectedRealmTuple `json:"tuple"`
	ItemIds blizzardv2.ItemIds                   `json:"item_ids"`
}

type GetPricelistResponse struct {
	Pricelist sotah.ItemPrices `json:"price_list"`
}

func (plResponse GetPricelistResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(plResponse)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (ladBases LiveAuctionsDatabases) GetPricelist(
	data []byte,
) (GetPricelistResponse, codes.Code, error) {
	plRequest, err := NewGetPricelistRequest(data)
	if err != nil {
		return GetPricelistResponse{}, codes.MsgJSONParseError, err
	}

	ladBase, err := ladBases.GetDatabase(plRequest.Tuple)
	if err != nil {
		return GetPricelistResponse{}, codes.UserError, err
	}

	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return GetPricelistResponse{}, codes.GenericError, err
	}

	iPrices := sotah.NewItemPricesFromMiniAuctionList(maList).FilterIn(plRequest.ItemIds)

	return GetPricelistResponse{Pricelist: iPrices}, codes.Ok, nil
}
