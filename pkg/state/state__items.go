package state

import (
	ItemsDatabase "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database/items" // nolint:lll
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

	if err := itemsDatabase.ResetItems(); err != nil {
		return ItemsState{}, err
	}

	hasItemClasses, err := itemsDatabase.HasItemClasses()
	if err != nil {
		logging.WithField(
			"error",
			err.Error(),
		).Error("failed to check if items-database has item-classes")

		return ItemsState{}, err
	}
	if !hasItemClasses {
		encodedItemClasses, err := opts.LakeClient.GetEncodedItemClasses()
		if err != nil {
			logging.WithField("error", err.Error()).Error("failed to gather encoded item-classes")

			return ItemsState{}, err
		}

		if err := itemsDatabase.PersistItemClasses(encodedItemClasses); err != nil {
			logging.WithField("error", err.Error()).Error("failed to persist encoded item-classes")

			return ItemsState{}, err
		}
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
		subjects.Items:                    sta.ListenForItems,
		subjects.ItemsQuery:               sta.ListenForItemsQuery,
		subjects.ItemsIntake:              sta.ListenForItemsIntake,
		subjects.ItemsFindMatchingRecipes: sta.ListenForItemsFindMatchingRecipes,
		subjects.ItemClasses:              sta.ListenForItemClasses,
		subjects.ItemSubjectsByItemClass:  sta.ListenForItemSubjectsByItemClass,
	}
}
