package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/database"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type ItemsState struct {
	BlizzardState BlizzardState
	RegionsState  *RegionsState

	Messenger     messenger.Messenger
	ItemsDatabase database.ItemsDatabase
}

func (sta ItemsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Items:      sta.ListenForItems,
		subjects.ItemsQuery: sta.ListenForItemsQuery,
	}
}
