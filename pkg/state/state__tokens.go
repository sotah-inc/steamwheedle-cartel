package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"

	TokensDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/tokens"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewTokensStateOptions struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	TokensDatabaseDir string
	Regions           sotah.RegionList
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
		Regions:        opts.Regions,
	}, nil
}

type TokensState struct {
	BlizzardState BlizzardState

	Messenger      messenger.Messenger
	TokensDatabase TokensDatabase.Database
	Reporter       metric.Reporter

	Regions sotah.RegionList
}

func (sta TokensState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.RegionTokenHistory: sta.ListenForRegionTokenHistory,
		subjects.TokenHistoryIntake: sta.ListenForTokenHistoryIntake,
	}
}
