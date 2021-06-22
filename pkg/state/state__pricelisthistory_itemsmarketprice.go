package state

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewItemsMarketPriceRequest(data []byte) (ItemsMarketPriceRequest, error) {
	req := &ItemsMarketPriceRequest{}
	err := json.Unmarshal(data, &req)
	if err != nil {
		return ItemsMarketPriceRequest{}, err
	}

	return *req, nil
}

type ItemsMarketPriceRequest struct {
	Tuple   blizzardv2.RegionVersionConnectedRealmTuple `json:"tuple"`
	ItemIds blizzardv2.ItemIds                          `json:"item_ids"`
}

type ItemsMarketPriceResponse struct {
	ItemsMarketPrice map[blizzardv2.ItemId]float64 `json:"items_market_price"`
}

func (res ItemsMarketPriceResponse) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(res)
}

func (sta PricelistHistoryState) ListenForItemsMarketPrice(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.ItemsMarketPrice), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		req, err := NewItemsMarketPriceRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		itemsMarketPrice, err := sta.PricelistHistoryDatabases.GetItemsMarketPrice(req.Tuple, req.ItemIds)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		res := ItemsMarketPriceResponse{ItemsMarketPrice: itemsMarketPrice}

		data, err := res.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = string(data)

		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
