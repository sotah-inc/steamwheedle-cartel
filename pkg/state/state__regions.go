package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

type NewRegionStateOptions struct {
	BlizzardState BlizzardState
	Regions       sotah.RegionList
	Messenger     messenger.Messenger
}

func NewRegionState(opts NewRegionStateOptions) (RegionsState, error) {
	regionConnectedRealms, err := opts.BlizzardState.ResolveRegionConnectedRealms(opts.Regions)
	if err != nil {
		return RegionsState{}, err
	}

	regionComposites := make(sotah.RegionComposites, len(opts.Regions))
	for i, region := range opts.Regions {
		connectedRealms := regionConnectedRealms[region.Name]

		realmComposites := make([]sotah.RealmComposite, len(connectedRealms))
		for j, response := range connectedRealms {
			realmComposites[j] = sotah.RealmComposite{ConnectedRealmResponse: response}
		}

		regionComposites[i] = sotah.RegionComposite{
			ConfigRegion:             region,
			ConnectedRealmComposites: realmComposites,
		}
	}

	return RegionsState{
		BlizzardState:    opts.BlizzardState,
		Messenger:        opts.Messenger,
		RegionComposites: &regionComposites,
	}, nil
}

type RegionsState struct {
	BlizzardState BlizzardState
	Messenger     messenger.Messenger

	RegionComposites *sotah.RegionComposites
}

func (sta RegionsState) ReceiveTimestamps(timestamps sotah.RegionTimestamps) {
	sta.RegionComposites = sta.RegionComposites.Receive(timestamps)
}

func (sta RegionsState) GetListeners() SubjectListeners {
	return SubjectListeners{
		subjects.Status:              sta.ListenForStatus,
		subjects.ValidateRegionRealm: sta.ListenForValidateRegionRealm,
	}
}
