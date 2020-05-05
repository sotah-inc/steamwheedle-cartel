package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	BaseLake "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/lake/base"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewItemsStateOptions struct {
	LakeClient BaseLake.Client
	Messenger  messenger.Messenger

	ItemsDatabaseDir string
}

func NewItemsState(opts NewItemsStateOptions) (ItemsState, error) {
	itemsDatabase, err := database.NewItemsDatabase(opts.ItemsDatabaseDir)
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
	ItemsDatabase database.ItemsDatabase
}

func (sta ItemsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Items:      sta.ListenForItems,
		subjects.ItemsQuery: sta.ListenForItemsQuery,
	}
}
