package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewRegionStateOptions struct {
	BlizzardState            BlizzardState
	Regions                  sotah.RegionList
	Messenger                messenger.Messenger
	RegionRealmSlugWhitelist sotah.RegionRealmSlugWhitelist
}

func NewRegionState(opts NewRegionStateOptions) (*RegionsState, error) {
	regionConnectedRealms, err := opts.BlizzardState.ResolveRegionConnectedRealms(opts.Regions)
	if err != nil {
		return nil, err
	}

	logging.WithField("whitelist", opts.RegionRealmSlugWhitelist).Info("checking with whitelist")

	regionComposites := make(sotah.RegionComposites, len(opts.Regions))
	for i, region := range opts.Regions {
		connectedRealms := regionConnectedRealms[region.Name]

		var realmComposites []sotah.RealmComposite
		for _, response := range connectedRealms {
			realmComposite := sotah.NewRealmComposite(opts.RegionRealmSlugWhitelist.Get(region.Name), response)
			if realmComposite.IsZero() {
				continue
			}

			realmComposites = append(realmComposites, realmComposite)
		}

		regionComposites[i] = sotah.RegionComposite{
			ConfigRegion:             region,
			ConnectedRealmComposites: realmComposites,
		}
	}

	return &RegionsState{
		BlizzardState:    opts.BlizzardState,
		Messenger:        opts.Messenger,
		RegionComposites: regionComposites,
	}, nil
}

type RegionsState struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	RegionComposites sotah.RegionComposites
}

func (sta *RegionsState) ReceiveTimestamps(timestamps sotah.RegionTimestamps) {
	result := sta.RegionComposites.Receive(timestamps)
	sta.RegionComposites = result
}

func (sta *RegionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Status:                       sta.ListenForStatus,
		subjects.ValidateRegionConnectedRealm: sta.ListenForValidateRegionConnectedRealm,
		subjects.QueryRealmModificationDates:  sta.ListenForQueryRealmModificationDates,
	}
}
