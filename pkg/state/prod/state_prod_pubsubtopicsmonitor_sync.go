package prod

import (
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"google.golang.org/api/iterator"
)

func (sta PubsubTopicsMonitorState) Sync() error {
	startTime := time.Now()
	totalTopicNames := 0
	it := sta.IO.BusClient.TopicNames()
	for {
		next, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return err
		}

		logging.WithField("topic", next.String()).Info("Found topic")
		totalTopicNames += 1
	}

	sta.IO.Reporter.Report(metric.Metrics{
		"pubsub_topics_monitor_sync_duration": int(int64(time.Since(startTime)) / 1000 / 1000 / 1000),
		"pubsub_topics_monitor_topic_count":   totalTopicNames,
	})

	return nil
}
