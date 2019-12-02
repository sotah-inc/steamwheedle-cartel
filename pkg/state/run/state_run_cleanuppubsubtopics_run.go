package run

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/sotah"
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
