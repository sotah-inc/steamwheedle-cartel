package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta CleanupManifestsState) Handle(regionRealmTuple sotah.RegionRealmTuple) sotah.Message {
	m := sotah.NewMessage()

	return m
}

func (sta CleanupManifestsState) Run(data []byte) sotah.Message {
	regionRealmTuple, err := sotah.NewRegionRealmTuple(string(data))
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	return sta.Handle(regionRealmTuple)
}
