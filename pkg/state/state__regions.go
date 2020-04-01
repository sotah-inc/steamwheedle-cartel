package state

import "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"

func NewRegionState(blizzardState BlizzardState, regions sotah.RegionList) (RegionsState, error) {
	regionConnectedRealms, err := blizzardState.ResolveRegionConnectedRealms(regions)
	if err != nil {
		return RegionsState{}, err
	}

	regionComposites := make(sotah.RegionComposites, len(regions))
	for i, region := range regions {
		regionComposites[i] = sotah.RegionComposite{
			Region:          region,
			ConnectedRealms: regionConnectedRealms[region.Name],
		}
	}

	return RegionsState{blizzardState, regionComposites}, nil
}

type RegionsState struct {
	BlizzardState BlizzardState

	RegionComposites sotah.RegionComposites
}
