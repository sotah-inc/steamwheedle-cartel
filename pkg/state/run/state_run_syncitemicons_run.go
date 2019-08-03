package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta SyncItemIconsState) Run(payloads sotah.IconItemsPayloads) sotah.Message {
	logging.WithField("payloads", payloads).Info("Handling")

	return sotah.NewMessage()
}
