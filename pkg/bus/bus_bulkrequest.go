package bus

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/twinj/uuid"
)

type MessageResponses struct {
	Items BulkRequestMessages
	Mutex *sync.Mutex
}

func (r MessageResponses) IsComplete() bool {
	for _, msg := range r.Items {
		if len(msg.ReplyToId) == 0 {
			return false
		}
	}

	return true
}

func (r MessageResponses) FilterInCompleted() BulkRequestMessages {
	out := BulkRequestMessages{}
	for k, v := range r.Items {
		if len(v.ReplyToId) == 0 {
			continue
		}

		out[k] = v
	}

	return out
}

func NewBulkRequestMessages(messages []Message) BulkRequestMessages {
	out := BulkRequestMessages{}
	for _, msg := range messages {
		out[msg.ReplyToId] = NewMessage()
	}

	return out
}

type BulkRequestMessages map[string]Message

func (c Client) BulkRequest(
	intakeTopic *pubsub.Topic,
	messages []Message,
	timeout time.Duration,
) (BulkRequestMessages, error) {
	// producing a topic to receive responses
	logging.Info("Producing a topic and subscription to receive responses")
	recipientTopic, err := c.CreateTopic(fmt.Sprintf("bulk-request-%s", uuid.NewV4().String()))
	if err != nil {
		return BulkRequestMessages{}, err
	}

	// updating messages with reply-to topic
	for i, msg := range messages {
		msg.ReplyTo = recipientTopic.ID()
		messages[i] = msg
	}

	// producing a blank list of message responses
	responses := MessageResponses{
		Mutex: &sync.Mutex{},
		Items: NewBulkRequestMessages(messages),
	}

	// opening a listener
	logging.Info("Opening a listener and waiting for it to finish opening")
	onComplete := make(chan interface{})
	receiveConfig := SubscribeConfig{
		Topic:     recipientTopic,
		OnReady:   make(chan interface{}),
		Stop:      make(chan interface{}),
		OnStopped: make(chan interface{}),
		Callback: func(busMsg Message) {
			responses.Mutex.Lock()
			defer responses.Mutex.Unlock()
			responses.Items[busMsg.ReplyToId] = busMsg

			if !responses.IsComplete() {
				return
			}

			onComplete <- struct{}{}
		},
	}
	go func() {
		if err := c.Subscribe(receiveConfig); err != nil {
			logging.Fatalf("Failed to subscribe to recipient topic: %s", err.Error())

			return
		}
	}()
	<-receiveConfig.OnReady

	// bulk publishing
	logging.Info("Bulk publishing")
	startTime := time.Now()
	for outJob := range c.BulkPublish(intakeTopic, messages) {
		if outJob.Err != nil {
			return BulkRequestMessages{}, outJob.Err
		}
	}

	// waiting for responses is complete or timer runs out
	logging.Info("Waiting for responses to complete or timer runs out")
	timer := time.After(timeout)
	select {
	case <-timer:
		logging.Info("Timer timed out, going over results in allotted time")

		break
	case <-onComplete:
		logging.Info("Received all responses, going over all responses")

		break
	}
	responseItems := responses.FilterInCompleted()
	duration := time.Since(startTime)

	// stopping the receiver
	logging.WithFields(
		logrus.Fields{
			"duration":  int(duration.Seconds()),
			"responses": len(responseItems),
		},
	).Info("Finished receiving responses, stopping the listener and waiting for it to stop")
	receiveConfig.Stop <- struct{}{}
	<-receiveConfig.OnStopped

	if err := recipientTopic.Delete(c.context); err != nil {
		return BulkRequestMessages{}, err
	}

	return responses.FilterInCompleted(), nil
}

func NewRegionRealmTimestampTuplesFromMessages(messages BulkRequestMessages) (RegionRealmTimestampTuples, error) {
	tuples := RegionRealmTimestampTuples{}
	for _, msg := range messages {
		tuple, err := NewRegionRealmTimestampTuple(msg.Data)
		if err != nil {
			return RegionRealmTimestampTuples{}, err
		}

		tuples = append(tuples, tuple)
	}

	return tuples, nil
}

func NewItemIdsFromMessages(messages BulkRequestMessages) (blizzard.ItemIds, error) {
	itemIdsMap := map[blizzard.ItemID]interface{}{}
	for _, msg := range messages {
		itemIds, err := blizzard.NewItemIds(msg.Data)
		if err != nil {
			return blizzard.ItemIds{}, err
		}

		for _, id := range itemIds {
			itemIdsMap[id] = struct{}{}
		}
	}

	out := blizzard.ItemIds{}
	for id := range itemIdsMap {
		out = append(out, id)
	}

	return out, nil
}

func NewPricelistHistoriesComputeIntakeRequestsFromMessages(
	messages BulkRequestMessages,
) (database.PricelistHistoriesComputeIntakeRequests, error) {
	out := database.PricelistHistoriesComputeIntakeRequests{}
	for _, msg := range messages {
		var respData database.PricelistHistoriesComputeIntakeRequest
		if err := json.Unmarshal([]byte(msg.Data), &respData); err != nil {
			return database.PricelistHistoriesComputeIntakeRequests{}, err
		}

		out = append(out, respData)
	}

	return out, nil
}
