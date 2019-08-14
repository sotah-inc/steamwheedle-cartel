package bus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
	"github.com/twinj/uuid"
	"google.golang.org/api/iterator"
)

func NewClient(projectID string, subscriberId string) (Client, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return Client{}, err
	}

	return Client{
		client:       client,
		context:      ctx,
		projectId:    projectID,
		subscriberId: subscriberId,
	}, nil
}

type Client struct {
	context      context.Context
	projectId    string
	client       *pubsub.Client
	subscriberId string
}

func (c Client) CreateTopic(id string) (*pubsub.Topic, error) {
	return c.client.CreateTopic(c.context, id)
}

func (c Client) Topic(topicName string) *pubsub.Topic {
	return c.client.Topic(topicName)
}

func (c Client) Topics() *pubsub.TopicIterator {
	return c.client.Topics(c.context)
}

func (c Client) Subscriptions(t *pubsub.Topic) *pubsub.SubscriptionIterator {
	return t.Subscriptions(c.context)
}

func (c Client) FirmTopic(topicName string) (*pubsub.Topic, error) {
	topic := c.Topic(topicName)

	exists, err := topic.Exists(c.context)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("topic does not exist")
	}

	return topic, nil
}

func (c Client) ResolveTopic(id string) (*pubsub.Topic, error) {
	topic := c.Topic(id)
	exists, err := topic.Exists(c.context)
	if err != nil {
		return nil, err
	}
	if exists {
		return topic, nil
	}

	return c.CreateTopic(id)
}

func (c Client) CreateSubscription(topic *pubsub.Topic) (*pubsub.Subscription, error) {
	return c.client.CreateSubscription(c.context, c.subscriberName(topic), pubsub.SubscriptionConfig{
		Topic: topic,
	})
}

func (c Client) Publish(topic *pubsub.Topic, msg Message) (string, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return topic.Publish(c.context, &pubsub.Message{Data: data}).Get(c.context)
}

type BulkPublishOutJob struct {
	Err error
	Msg Message
}

func (c Client) BulkPublish(topic *pubsub.Topic, messages []Message) chan BulkPublishOutJob {
	// opening workers and channels
	in := make(chan Message)
	out := make(chan BulkPublishOutJob)
	worker := func() {
		for msg := range in {
			if _, err := c.Publish(topic, msg); err != nil {
				out <- BulkPublishOutJob{
					Err: err,
					Msg: msg,
				}

				continue
			}

			out <- BulkPublishOutJob{
				Err: nil,
				Msg: msg,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(32, worker, postWork)

	// queueing it up
	go func() {
		for _, msg := range messages {
			in <- msg
		}

		close(in)
	}()

	return out
}

func (c Client) subscriberName(topic *pubsub.Topic) string {
	return fmt.Sprintf("subscriber-%s-%s-%s", c.subscriberId, topic.ID(), uuid.NewV4().String())
}

func (c Client) SubscribeToTopic(id string, config SubscribeConfig) error {
	topic, err := c.ResolveTopic(id)
	if err != nil {
		return err
	}
	config.Topic = topic

	return c.Subscribe(config)
}

type SubscribeConfig struct {
	Topic     *pubsub.Topic
	Stop      chan interface{}
	OnReady   chan interface{}
	OnStopped chan interface{}
	Callback  func(Message)
}

func (c Client) Subscribe(config SubscribeConfig) error {
	sub, err := c.CreateSubscription(config.Topic)
	if err != nil {
		return err
	}

	config.OnReady <- struct{}{}

	entry := logging.WithFields(logrus.Fields{
		"subscriber-name": sub.ID(),
		"topic":           config.Topic.ID(),
	})

	cctx, cancel := context.WithCancel(c.context)
	go func() {
		<-config.Stop

		cancel()
		config.Topic.Stop()

		config.OnStopped <- struct{}{}
	}()

	entry.Info("Waiting for messages")
	err = sub.Receive(cctx, func(ctx context.Context, pubsubMsg *pubsub.Message) {
		pubsubMsg.Ack()

		var msg Message
		if err := json.Unmarshal(pubsubMsg.Data, &msg); err != nil {
			entry.WithField("error", err.Error()).Error("Failed to parse message")

			return
		}

		config.Callback(msg)
	})
	if err != nil {
		if err == context.Canceled {
			return nil
		}

		return err
	}

	return nil
}

func (c Client) ReplyToWithError(recipient Message, err error, code codes.Code) error {
	reply := NewMessage()
	reply.Code = code
	reply.Err = err.Error()
	if _, err := c.ReplyTo(recipient, reply); err != nil {
		return err
	}

	return nil
}

func (c Client) ReplyTo(target Message, payload Message) (string, error) {
	if target.ReplyTo == "" {
		return "", errors.New("cannot reply to blank reply-to topic name")
	}

	// validating topic already exists
	topic, err := c.FirmTopic(target.ReplyTo)
	if err != nil {
		return "", err
	}

	logging.WithField("reply-to-topic", topic.ID()).Info("Replying to topic")

	return c.Publish(topic, payload)
}

func (c Client) RequestFromTopic(topicName string, payload string, timeout time.Duration) (Message, error) {
	topic, err := c.FirmTopic(topicName)
	if err != nil {
		return Message{}, err
	}

	return c.Request(topic, payload, timeout)
}

type requestJob struct {
	Err     error
	Payload Message
}

func (c Client) Request(recipientTopic *pubsub.Topic, payload string, timeout time.Duration) (Message, error) {
	// producing a reply-to topic
	replyToTopic, err := c.client.CreateTopic(c.context, fmt.Sprintf("reply-to-%s", uuid.NewV4().String()))
	if err != nil {
		return Message{}, err
	}

	// producing a reply-to subscription
	replyToSub, err := c.client.CreateSubscription(c.context, c.subscriberName(replyToTopic), pubsub.SubscriptionConfig{
		Topic: replyToTopic,
	})
	if err != nil {
		return Message{}, err
	}

	cctx, cancel := context.WithCancel(c.context)

	// spawning a worker to wait for a response on the reply-to topic
	out := make(chan requestJob)
	go func() {
		// spawning a receiver worker to receive the results and push them out
		receiver := make(chan requestJob)
		go func() {
			select {
			case result := <-receiver:
				close(receiver)

				cancel()
				if err := replyToSub.Delete(c.context); err != nil {
					logging.WithFields(logrus.Fields{
						"error":        err.Error(),
						"subscription": replyToSub.ID(),
					}).Error("Failed to delete reply-to subscription after receiving result")

					out <- requestJob{
						Err:     err,
						Payload: Message{},
					}

					return
				}

				replyToTopic.Stop()
				if err := replyToTopic.Delete(c.context); err != nil {
					logging.WithFields(logrus.Fields{
						"error": err.Error(),
						"topic": replyToTopic.ID(),
					}).Error("Failed to delete reply-to topic after receiving result")

					return
				}

				out <- result

				return
			case <-time.After(timeout):
				close(receiver)

				cancel()
				if err := replyToSub.Delete(c.context); err != nil {
					logging.WithFields(logrus.Fields{
						"error":        err.Error(),
						"subscription": replyToSub.ID(),
					}).Error("Failed to delete reply-to subscription after timing out")

					out <- requestJob{
						Err:     err,
						Payload: Message{},
					}

					return
				}

				replyToTopic.Stop()
				if err := replyToTopic.Delete(c.context); err != nil {
					logging.WithFields(logrus.Fields{
						"error": err.Error(),
						"topic": replyToTopic.ID(),
					}).Error("Failed to delete reply-to topic after timing out")

					out <- requestJob{
						Err:     err,
						Payload: Message{},
					}

					return
				}

				out <- requestJob{
					Err:     errors.New("timed out"),
					Payload: Message{},
				}

				return
			}
		}()

		// waiting for a message to come through
		err = replyToSub.Receive(cctx, func(ctx context.Context, pubsubMsg *pubsub.Message) {
			pubsubMsg.Ack()

			var msg Message
			if err := json.Unmarshal(pubsubMsg.Data, &msg); err != nil {
				receiver <- requestJob{
					Err:     err,
					Payload: Message{},
				}

				return
			}

			receiver <- requestJob{
				Err:     nil,
				Payload: msg,
			}
		})
		if err != nil {
			if err == context.Canceled {
				return
			}

			close(receiver)
			cancel()
			replyToTopic.Stop()

			out <- requestJob{
				Err:     err,
				Payload: Message{},
			}

			return
		}
	}()

	// publishing the payload to the recipient topic
	msg := NewMessage()
	msg.Data = payload
	msg.ReplyTo = replyToTopic.ID()
	jsonEncodedMessage, err := json.Marshal(msg)
	if err != nil {
		close(out)

		return Message{}, err
	}

	if _, err := recipientTopic.Publish(c.context, &pubsub.Message{Data: jsonEncodedMessage}).Get(c.context); err != nil {
		close(out)

		return Message{}, err
	}

	// waiting for a result to come out
	requestResult := <-out

	close(out)

	if requestResult.Err != nil {
		return Message{}, requestResult.Err
	}

	return requestResult.Payload, nil
}

func (c Client) PublishMetrics(m metric.Metrics) error {
	topic, err := c.FirmTopic(string(subjects.AppMetrics))
	if err != nil {
		return err
	}

	jsonEncoded, err := json.Marshal(m)
	if err != nil {
		return err
	}

	msg := NewMessage()
	msg.Data = string(jsonEncoded)
	if _, err := c.Publish(topic, msg); err != nil {
		return err
	}

	return nil
}

type PruneTopicsOutJob struct {
	Err     error
	Name    string
	Existed bool
}

func (job PruneTopicsOutJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"name":  job.Name,
	}
}

func NewPruneTopicsResults(data []byte) (PruneTopicsResults, error) {
	var out PruneTopicsResults
	if err := json.Unmarshal(data, &out); err != nil {
		return PruneTopicsResults{}, err
	}

	return out, nil
}

type PruneTopicsResults struct {
	Pruned      []string
	NonExistent []string
	Erroneous   []string
}

func (results PruneTopicsResults) EncodeForDelivery() ([]byte, error) {
	out, err := json.Marshal(results)
	if err != nil {
		return []byte{}, err
	}

	return out, nil
}

func (c Client) PruneTopics(names []string) PruneTopicsResults {
	// opening workers and channels
	in := make(chan string)
	out := make(chan PruneTopicsOutJob)
	worker := func() {
		for name := range in {
			topic := c.Topic(name)

			exists, err := topic.Exists(c.context)
			if err != nil {
				out <- PruneTopicsOutJob{
					Err:  err,
					Name: name,
				}

				continue
			}
			if !exists {
				out <- PruneTopicsOutJob{
					Err:     nil,
					Name:    name,
					Existed: false,
				}

				continue
			}

			if err := topic.Delete(c.context); err != nil {
				out <- PruneTopicsOutJob{
					Err:  err,
					Name: name,
				}

				continue
			}

			out <- PruneTopicsOutJob{
				Err:     nil,
				Name:    name,
				Existed: true,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(16, worker, postWork)

	// queueing it up
	go func() {
		for _, name := range names {
			in <- name
		}

		close(in)
	}()

	// waiting for it to drain out
	results := PruneTopicsResults{
		Pruned:      []string{},
		NonExistent: []string{},
		Erroneous:   []string{},
	}
	for outJob := range out {
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to prune topic")

			results.Erroneous = append(results.Erroneous, outJob.Name)

			continue
		}

		if !outJob.Existed {
			results.NonExistent = append(results.NonExistent, outJob.Name)

			continue
		}

		results.Pruned = append(results.Pruned, outJob.Name)
	}

	return results
}

type CheckSubscriptionsResult struct {
	TopicName        string
	HasSubscriptions bool
}

type CheckSubscriptionsResults []CheckSubscriptionsResult

func (r CheckSubscriptionsResults) TopicNames() []string {
	out := make([]string, len(r))
	for i, result := range r {
		out[i] = result.TopicName
	}

	return out
}

func (r CheckSubscriptionsResults) WithoutSubscriptions() CheckSubscriptionsResults {
	out := CheckSubscriptionsResults{}
	for _, result := range r {
		if result.HasSubscriptions {
			continue
		}

		out = append(out, result)
	}

	return out
}

type CheckSubscriptionsOutJob struct {
	Err              error
	TopicName        string
	HasSubscriptions bool
}

func (c Client) CheckAllSubscriptions(maxCount int) (CheckSubscriptionsResults, error) {
	// opening workers and channels
	in := make(chan string)
	out := make(chan CheckSubscriptionsOutJob)
	worker := func() {
		for topicName := range in {
			topic := c.Topic(topicName)

			hasSubscriptions, err := func() (bool, error) {
				subsIterator := topic.Subscriptions(c.context)
				if _, err := subsIterator.Next(); err != nil {
					if err == iterator.Done {
						return false, nil
					}

					return false, err
				}

				return true, nil
			}()
			if err != nil {
				out <- CheckSubscriptionsOutJob{
					Err:       err,
					TopicName: topicName,
				}

				continue
			}

			logging.WithField("topic-name", topicName).Info("Checked")

			out <- CheckSubscriptionsOutJob{
				Err:              nil,
				TopicName:        topicName,
				HasSubscriptions: hasSubscriptions,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(8, worker, postWork)

	// queueing it up
	go func() {
		i := 0
		it := c.Topics()
		for {
			next, err := it.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}

				logging.WithField("error", err.Error()).Error("Failed to iterate to next topic")

				continue
			}

			in <- next.ID()

			i++
			if i >= maxCount {
				break
			}
		}

		close(in)
	}()

	// waiting for it to drain out
	results := CheckSubscriptionsResults{}
	for outJob := range out {
		if outJob.Err != nil {
			return CheckSubscriptionsResults{}, outJob.Err
		}

		results = append(results, CheckSubscriptionsResult{
			TopicName:        outJob.TopicName,
			HasSubscriptions: outJob.HasSubscriptions,
		})
	}

	return results, nil
}
