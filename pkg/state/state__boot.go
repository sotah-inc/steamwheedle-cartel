package state

import (
	"github.com/twinj/uuid"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewBootStateOptions struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	Regions     sotah.RegionList
	Expansions  []sotah.Expansion
	Professions []sotah.Profession
}

func NewBootState(opts NewBootStateOptions) (BootState, error) {
	itemClasses, err := opts.BlizzardState.ResolveItemClasses(opts.Regions)
	if err != nil {
		return BootState{}, err
	}

	return BootState{
		Regions:       opts.Regions,
		Messenger:     opts.Messenger,
		SessionSecret: uuid.NewV4(),
		ItemClasses:   itemClasses,
		Expansions:    opts.Expansions,
		Professions: func() []sotah.Profession {
			out := make([]sotah.Profession, len(opts.Professions))
			for i, prof := range opts.Professions {
				prof.IconURL = blizzardv2.DefaultGetItemIconURL(opts.Professions[i].Icon)
				out[i] = prof
			}

			return out
		}(),
	}, nil
}

type BootState struct {
	Messenger messenger.Messenger

	// initialized at runtime
	SessionSecret uuid.UUID
	ItemClasses   []blizzardv2.ItemClassResponse

	// receiving from config file
	Regions     sotah.RegionList
	Expansions  []sotah.Expansion
	Professions []sotah.Profession
}

func (sta BootState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Boot:          sta.ListenForBoot,
		subjects.SessionSecret: sta.ListenForSessionSecret,
	}
}
