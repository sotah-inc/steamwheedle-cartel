package run

import (
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/logging"
)

func (sta GatewayState) CleanupAllPubsubTopics() error {
	logging.Info("Starting cleanup-all-pubsub-topics")

	logging.Info("Finished")

	return nil
}
