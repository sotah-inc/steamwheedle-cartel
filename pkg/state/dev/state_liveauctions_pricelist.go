package dev

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"

	nats "github.com/nats-io/go-nats"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

func newPriceListRequest(payload []byte) (priceListRequest, error) {
	pList := &priceListRequest{}
	err := json.Unmarshal(payload, &pList)
	if err != nil {
		return priceListRequest{}, err
	}

	return *pList, nil
}

type priceListRequest struct {
	RegionName blizzard.RegionName `json:"region_name"`
	RealmSlug  blizzard.RealmSlug  `json:"realm_slug"`
	ItemIds    []blizzard.ItemID   `json:"item_ids"`
}

func (plRequest priceListRequest) resolve(laState LiveAuctionsState) (sotah.MiniAuctionList, state.RequestError) {
	regionLadBases, ok := laState.IO.Databases.LiveAuctionsDatabases[plRequest.RegionName]
	if !ok {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.NotFound, Message: "Invalid region"}
	}

	ladBase, ok := regionLadBases[plRequest.RealmSlug]
	if !ok {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.NotFound, Message: "Invalid realm"}
	}

	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return sotah.MiniAuctionList{}, state.RequestError{Code: codes.GenericError, Message: err.Error()}
	}

	return maList, state.RequestError{Code: codes.Ok, Message: ""}
}

type priceListResponse struct {
	PriceList sotah.ItemPrices `json:"price_list"`
}

func (plResponse priceListResponse) encodeForMessage() (string, error) {
	jsonEncodedMessage, err := json.Marshal(plResponse)
	if err != nil {
		return "", err
	}

	gzipEncodedMessage, err := util.GzipEncode(jsonEncodedMessage)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncodedMessage), nil
}

func (laState LiveAuctionsState) ListenForPriceList(stop state.ListenStopChan) error {
	err := laState.IO.Messenger.Subscribe(string(subjects.PriceList), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		plRequest, err := newPriceListRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// resolving data from state
		realmAuctions, reErr := plRequest.resolve(laState)
		if reErr.Code != codes.Ok {
			m.Err = reErr.Message
			m.Code = reErr.Code
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// deriving a pricelist-response from the provided realm auctions
		iPrices := sotah.NewItemPrices(realmAuctions)
		responseItemPrices := sotah.ItemPrices{}
		for _, itemId := range plRequest.ItemIds {
			if iPrice, ok := iPrices[itemId]; ok {
				responseItemPrices[itemId] = iPrice

				continue
			}
		}

		plResponse := priceListResponse{responseItemPrices}
		data, err := plResponse.encodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			laState.IO.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		laState.IO.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
