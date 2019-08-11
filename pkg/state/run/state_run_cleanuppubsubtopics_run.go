package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta CleanupPubsubTopicsState) Run(topicNames []string) sotah.Message {
	encodedResults, err := sta.IO.BusClient.PruneTopics(topicNames).EncodeForDelivery()
	if err != nil {
		return sotah.NewErrorMessage(err)
	}

	out := sotah.NewMessage()
	out.Data = string(encodedResults)

	return out
}
