package state

import (
	"encoding/base64"
	"encoding/json"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemsRequest(payload []byte) (ItemsRequest, error) {
	iRequest := &ItemsRequest{}
	err := json.Unmarshal(payload, &iRequest)
	if err != nil {
		return ItemsRequest{}, err
	}

	return *iRequest, nil
}

type ItemsRequest struct {
	Locale  locale.Locale           `json:"locale"`
	Version gameversion.GameVersion `json:"game_version"`
	ItemIds blizzardv2.ItemIds      `json:"itemIds"`
}

type ItemsResponse struct {
	Items sotah.ShortItemList `json:"items"`
}

func (iResponse ItemsResponse) EncodeForMessage() (string, error) {
	encodedResult, err := json.Marshal(iResponse)
	if err != nil {
		return "", err
	}

	gzippedResult, err := util.GzipEncode(encodedResult)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzippedResult), nil
}

func (sta ItemsState) ListenForItems(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.Items), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		iRequest, err := NewItemsRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		if iRequest.Locale.IsZero() {
			m.Err = "locale was zero"
			m.Code = codes.UserError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		logging.WithField("request", iRequest).Info("received items request")

		itemsOut := sta.ItemsDatabase.FindItems(iRequest.Version, iRequest.ItemIds)
		var foundItems []sotah.Item
		for itemsOutJob := range itemsOut {
			if itemsOutJob.Err != nil {
				m.Err = itemsOutJob.Err.Error()
				m.Code = codes.GenericError
				sta.Messenger.ReplyTo(natsMsg, m)

				return
			}

			if !itemsOutJob.Exists {
				continue
			}

			foundItems = append(foundItems, itemsOutJob.Item)
		}

		resolvedShortItems := sotah.NewShortItemList(foundItems, iRequest.Locale)

		iResponse := ItemsResponse{Items: resolvedShortItems}
		data, err := iResponse.EncodeForMessage()
		if err != nil {
			m.Err = err.Error()
			m.Code = codes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		m.Data = data
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}
