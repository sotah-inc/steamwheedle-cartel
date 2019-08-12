package run

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func (sta GatewayState) CleanupAllPubsubTopics() error {
	logging.Info("Starting cleanup-all-pubsub-topics")

	logging.Info("Finished")

	return nil
}
