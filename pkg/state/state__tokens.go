package state

import (
	"encoding/json"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	TokensDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/tokens"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	mCodes "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewTokensStateOptions struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	TokensDatabaseDir string
}

func NewTokensState(opts NewTokensStateOptions) (TokensState, error) {
	if err := util.EnsureDirExists(opts.TokensDatabaseDir); err != nil {
		logging.WithField("error", err.Error()).Error("failed to ensure tokens-database-dir exists")

		return TokensState{}, err
	}

	tokensDatabase, err := TokensDatabase.NewDatabase(opts.TokensDatabaseDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise tokens-database")

		return TokensState{}, err
	}

	return TokensState{
		BlizzardState:  opts.BlizzardState,
		Messenger:      opts.Messenger,
		TokensDatabase: tokensDatabase,
		Reporter:       metric.NewReporter(opts.Messenger),
	}, nil
}

type TokensState struct {
	BlizzardState BlizzardState

	Messenger      messenger.Messenger
	TokensDatabase TokensDatabase.Database
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
	logging.Info("collecting region-tokens")

	startTime := time.Now()

	// gathering tokens
	tokens, err := sta.BlizzardState.ResolveTokens(regions)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to fetch tokens")

		return err
	}

	// formatting appropriately
	regionTokenHistory := TokensDatabase.RegionTokenHistory{}
	for regionName, token := range tokens {
		regionTokenHistory[regionName] = TokensDatabase.TokenHistory{token.LastUpdatedTimestamp: token.Price}
	}

	if err := sta.TokensDatabase.PersistHistory(regionTokenHistory); err != nil {
		logging.WithField("error", err.Error()).Error("failed to persist tokens-history")

		return err
	}
	logging.WithFields(logrus.Fields{
		"total":          len(tokens),
		"duration-in-ms": time.Since(startTime).Milliseconds(),
	}).Info("total persisted in collect-region-tokens")

	return nil
}
