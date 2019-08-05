package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta ComputePricelistHistoriesState) Run(tuple sotah.RegionRealmTimestampTuple) sotah.Message {
	logging.WithField("tuple", tuple).Info("Received tuple")

	return sotah.NewMessage()
}
