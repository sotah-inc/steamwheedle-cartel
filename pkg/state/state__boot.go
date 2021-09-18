package state

import (
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewBootStateOptions struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	GameVersionList gameversion.List
	Regions         sotah.RegionList
	Expansions      []sotah.Expansion
	FirebaseConfig  sotah.FirebaseConfig
}

func NewBootState(opts NewBootStateOptions) (BootState, error) {
	return BootState{
		GameVersionList: opts.GameVersionList,
		Regions:         opts.Regions,
		Messenger:       opts.Messenger,
		SessionSecret:   uuid.NewV4(),
		Expansions:      opts.Expansions,
		FirebaseConfig:  opts.FirebaseConfig,
	}, nil
}

type BootState struct {
	Messenger messenger.Messenger

	// initialized at runtime
	SessionSecret uuid.UUID

	// receiving from config file
	GameVersionList gameversion.List
	Regions         sotah.RegionList
	Expansions      []sotah.Expansion
	FirebaseConfig  sotah.FirebaseConfig
}

func (sta BootState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Boot:          sta.ListenForBoot,
		subjects.SessionSecret: sta.ListenForSessionSecret,
	}
}
