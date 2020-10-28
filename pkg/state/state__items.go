package state

import (
	ItemsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/items"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type NewItemsStateOptions struct {
	LakeClient BaseLake.Client
	Messenger  messenger.Messenger

	ItemsDatabaseDir string
}

func NewItemsState(opts NewItemsStateOptions) (ItemsState, error) {
	if err := util.EnsureDirExists(opts.ItemsDatabaseDir); err != nil {
		logging.WithField("error", err.Error()).Error("failed to ensure items-database-dir exists")

		return ItemsState{}, err
	}

	itemsDatabase, err := ItemsDatabase.NewDatabase(opts.ItemsDatabaseDir)
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to initialise items-database")

		return ItemsState{}, err
	}

	return ItemsState{
		LakeClient:    opts.LakeClient,
		Messenger:     opts.Messenger,
		ItemsDatabase: itemsDatabase,
	}, nil
}

type ItemsState struct {
	LakeClient    BaseLake.Client
	Messenger     messenger.Messenger
	ItemsDatabase ItemsDatabase.Database
}

func (sta ItemsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Items:       sta.ListenForItems,
		subjects.ItemsQuery:  sta.ListenForItemsQuery,
		subjects.ItemsIntake: sta.ListenForItemsIntake,
	}
}
