package blizzardv2

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type AuctionId int64

type AuctionItemModifier struct {
	Type  int `json:"type"`
	Value int `json:"value"`
}

type Auction struct {
	Id   AuctionId `json:"id"`
	Item struct {
		Id           ItemId                `json:"id"`
		Context      int                   `json:"context"`
		BonusLists   []int                 `json:"bonus_lists"`
		Modifiers    []AuctionItemModifier `json:"modifiers"`
		PetBreedId   int                   `json:"pet_breed_id"`
		PetLevel     int                   `json:"pet_level"`
		PetQualityId int                   `json:"pet_quality_id"`
		PetSpeciesId int                   `json:"pet_species_id"`
	} `json:"item"`
	Buyout   int64  `json:"buyout"`
	Quantity int    `json:"quantity"`
	TimeLeft string `json:"time_left"`
}

type AuctionsResponse struct {
	LinksBase
	ConnectedRealm HrefReference `json:"connected_realm"`
	Auctions       []Auction     `json:"auctions"`
}

func NewAuctionsFromHTTP(uri string) (AuctionsResponse, blizzard.ResponseMeta, error) {
	resp, err := blizzard.Download(uri)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to download auctions")

		return AuctionsResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    uri,
		}).Error("resp from auctions was not 200")

		return AuctionsResponse{}, resp, errors.New("status was not 200")
	}

	auctions, err := NewAuctionsResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   uri,
		}).Error("failed to parse auctions response")

		return AuctionsResponse{}, resp, err
	}

	return auctions, resp, nil
}

func NewAuctionsResponse(body []byte) (AuctionsResponse, error) {
	auctions := &AuctionsResponse{}
	if err := json.Unmarshal(body, auctions); err != nil {
		return AuctionsResponse{}, err
	}

	return *auctions, nil
}
