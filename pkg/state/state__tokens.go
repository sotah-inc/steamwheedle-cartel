package state

import (
	"encoding/json"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

	nats "github.com/nats-io/nats.go"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type TokensState struct {
	BlizzardState

	Messenger      messenger.Messenger
	TokensDatabase database.TokensDatabase
	Reporter       metric.Reporter
}

func (sta TokensState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.TokenHistory: sta.ListenForTokenHistory,
	}
}

func NewTokenHistoryRequest(data []byte) (TokenHistoryRequest, error) {
	var out TokenHistoryRequest
	if err := json.Unmarshal(data, &out); err != nil {
		return TokenHistoryRequest{}, err
	}

	return out, nil
}

type TokenHistoryRequest struct {
	RegionName string `json:"region_name"`
}

func (sta TokensState) ListenForTokenHistory(stop ListenStopChan) error {
	err := sta.Messenger.Subscribe(string(subjects.TokenHistory), stop, func(natsMsg nats.Msg) {
		m := messenger.NewMessage()

		// resolving the request
		request, err := NewTokenHistoryRequest(natsMsg.Data)
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// fetching token-history with request data
		tHistory, err := sta.TokensDatabase.GetHistory(blizzardv2.RegionName(request.RegionName))
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.MsgJSONParseError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// marshalling for messenger
		encodedMessage, err := tHistory.EncodeForDelivery()
		if err != nil {
			m.Err = err.Error()
			m.Code = mCodes.GenericError
			sta.Messenger.ReplyTo(natsMsg, m)

			return
		}

		// dumping it out
		m.Data = string(encodedMessage)
		sta.Messenger.ReplyTo(natsMsg, m)
	})
	if err != nil {
		return err
	}

	return nil
}

func (sta TokensState) CollectRegionTokens(regions sotah.RegionList) error {
	logging.Info("Collecting region-tokens")

	// gathering tokens
	tokens, err := sta.ResolveTokens(regions)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to fetch tokens")

		return err
	}

	// formatting appropriately
	regionTokenHistory := database.RegionTokenHistory{}
	for regionName, token := range tokens {
		regionTokenHistory[regionName] = database.TokenHistory{token.LastUpdatedTimestamp: token.Price}
	}

	return sta.TokensDatabase.PersistHistory(regionTokenHistory)
}
