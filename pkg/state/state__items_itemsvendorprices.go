package state

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func NewItemsVendorPricesRequest(data []byte) (ItemsVendorPricesRequest, error) {
	out := ItemsVendorPricesRequest{}
	if err := json.Unmarshal(data, &out); err != nil {
		return ItemsVendorPricesRequest{}, err
	}

	return out, nil
}

type ItemsVendorPricesRequest struct {
	Version gameversion.GameVersion `json:"game_version"`
	ItemIds blizzardv2.ItemIds      `json:"item_ids"`
}

type ItemsVendorPricesResponse struct {
	VendorPrices map[blizzardv2.ItemId]blizzardv2.PriceValue `json:"vendor_prices"`
}

func (res ItemsVendorPricesResponse) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(res)
	if err != nil {
		return "", err
	}

	return string(jsonEncoded), nil
}

func (sta ItemsState) ListenForItemsVendorPrices(stop ListenStopChan) error {
	return sta.Messenger.Subscribe(
		string(subjects.ItemsVendorPrices),
		stop,
		func(natsMsg nats.Msg) {
			m := messenger.NewMessage()

			logging.Info("handling request for items items-vendor-prices")

			req, err := NewItemsVendorPricesRequest(natsMsg.Data)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.MsgJSONParseError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			ivMap, err := sta.ItemsDatabase.VendorPrices(req.Version, req.ItemIds)
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			res := ItemsVendorPricesResponse{VendorPrices: ivMap}
			encodedData, err := res.EncodeForDelivery()
			if err != nil {
				m.Err = err.Error()
				m.Code = mCodes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			// dumping it out
			m.Data = encodedData
			sta.Messenger.ReplyTo(natsMsg, m)
		},
	)
}
