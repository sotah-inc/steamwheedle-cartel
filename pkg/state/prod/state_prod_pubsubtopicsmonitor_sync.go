package prod

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"google.golang.org/api/iterator"
)

func (sta PubsubTopicsMonitorState) Sync() error {
	startTime := time.Now()
	topicNames := []string{}
	it := sta.IO.BusClient.TopicNames()
	for {
		next, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return err
		}

		topicName := next.String()
		topicNames = append(topicNames, topicName)

		logging.WithField("topic", topicName).Info("Found topic")
	}

	currentSeen, err := sta.IO.Databases.PubsubTopicsDatabase.Fill(topicNames, time.Now())
	if err != nil {
		return err
	}

	retentionLimit := time.Now().Add(-1 * time.Hour * 24)

	logging.WithFields(logrus.Fields{
		"current-seen":    len(currentSeen.NonZero()),
		"total-seen":      len(currentSeen),
		"expired-seen":    len(currentSeen.After(retentionLimit)),
		"retention-limit": retentionLimit.String(),
	}).Info("Topic-names provided")

	sta.IO.Reporter.Report(metric.Metrics{
		"pubsub_topics_monitor_sync_duration": int(int64(time.Since(startTime)) / 1000 / 1000 / 1000),
		"pubsub_topics_monitor_topic_count":   len(topicNames),
	})

	return nil
}

func (sta PubsubTopicsMonitorState) ListenForSyncPubsubTopicsMonitor(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	in := make(chan interface{})
	go func() {
		for range in {
			if err := sta.Sync(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to call Sync()")

				continue
			}
		}
	}()

	// establishing subscriber config
	config := bus.SubscribeConfig{
		Stop: stop,
		Callback: func(busMsg bus.Message) {
			logging.WithField("bus-msg", busMsg).Info("Received bus-message")
			in <- struct{}{}
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := sta.IO.BusClient.SubscribeToTopic(string(subjects.SyncPubsubTopicsMonitor), config); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
