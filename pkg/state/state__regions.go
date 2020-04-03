package state

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

func NewRegionState(blizzardState BlizzardState, regions sotah.RegionList) (*RegionsState, error) {
	regionConnectedRealms, err := blizzardState.ResolveRegionConnectedRealms(regions)
	if err != nil {
		return nil, err
	}

	regionComposites := make(sotah.RegionComposites, len(regions))
	for i, region := range regions {
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

	return &RegionsState{blizzardState, regionComposites}, nil
}

type RegionsState struct {
	BlizzardState BlizzardState

	RegionComposites sotah.RegionComposites
}

func (sta RegionsState) ReceiveTimestamps(timestamps sotah.RegionTimestamps) {
	sta.RegionComposites.Receive(timestamps)
}
