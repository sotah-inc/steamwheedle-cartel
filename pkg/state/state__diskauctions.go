package state

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/diskstore"
)

type DiskAuctionsState struct {
	BlizzardState BlizzardState

	DiskStore diskstore.DiskStore
}

func (sta DiskAuctionsState) Collect(regionConnectedRealms blizzardv2.RegionConnectedRealmResponses) error {

}
