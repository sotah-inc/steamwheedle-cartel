package state

import (
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewBootStateOptions struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	Regions    sotah.RegionList
	Expansions []sotah.Expansion
}

func NewBootState(opts NewBootStateOptions) (BootState, error) {
	return BootState{
		Regions:       opts.Regions,
		Messenger:     opts.Messenger,
		SessionSecret: uuid.NewV4(),
		Expansions:    opts.Expansions,
	}, nil
}

type BootState struct {
	Messenger messenger.Messenger

	// initialized at runtime
	SessionSecret uuid.UUID

	// receiving from config file
	Regions    sotah.RegionList
	Expansions []sotah.Expansion
}

func (sta BootState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Boot:          sta.ListenForBoot,
		subjects.SessionSecret: sta.ListenForSessionSecret,
	}
}
