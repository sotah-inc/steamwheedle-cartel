package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta CleanupPubsubTopicsState) Run(topicNames []string) sotah.Message {
	return sotah.NewMessage()
}
