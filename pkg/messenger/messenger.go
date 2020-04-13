package messenger

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
)

type Messenger struct {
	Config Config
	conn   *nats.Conn
}

func NewMessage() Message {
	return Message{Code: codes.Ok}
}

type Message struct {
	Data string     `json:"data"`
	Err  string     `json:"error"`
	Code codes.Code `json:"code"`
}

func NewMessengerFromEnvVars(hostKey string, portKey string) (Messenger, error) {
	natsHost := os.Getenv(hostKey)
	natsPort := os.Getenv(portKey)
	if len(natsPort) == 0 {
		return Messenger{}, errors.New("nats port cannot be blank")
	}

	parsedNatsPort, err := strconv.Atoi(natsPort)
	if err != nil {
		return Messenger{}, err
	}

	return NewMessenger(Config{
		Hostname: natsHost,
		Port:     parsedNatsPort,
	})
}

type Config struct {
	Hostname string
	Port     int
}

func NewMessenger(config Config) (Messenger, error) {
	if len(config.Hostname) == 0 {
		return Messenger{}, errors.New("host cannot be blank")
	}

	if config.Port == 0 {
		return Messenger{}, errors.New("port cannot be zero")
	}

	natsURI := fmt.Sprintf("nats://%s:%d", config.Hostname, config.Port)

	logging.WithField("uri", natsURI).Info("Connecting to nats")

	conn, err := nats.Connect(natsURI)
	if err != nil {
		return Messenger{}, err
	}

	mess := Messenger{conn: conn, Config: config}

	return mess, nil
}

func (mess Messenger) Subscribe(subject string, stop chan interface{}, cb func(nats.Msg)) error {
	logging.WithField("subject", subject).Debug("Subscribing to subject")

	if mess.conn == nil {
		logging.WithField("config", mess.Config).Error("messenger connection was nil")

		return errors.New("messenger connection was nil")
	}

	sub, err := mess.conn.Subscribe(subject, func(natsMsg *nats.Msg) {
		logging.WithField("subject", subject).Debug("Received Request")

		cb(*natsMsg)
	})
	if err != nil {
		return err
	}

	go func() {
		<-stop

		logging.WithField("subject", subject).Info("Unsubscribing from subject")

		if err := sub.Unsubscribe(); err != nil {
			logging.WithField("error", err.Error()).Error("failed to unsubscribe")

			return
		}
	}()

	return nil
}

func (mess Messenger) ReplyTo(natsMsg nats.Msg, m Message) {
	if m.Code == codes.Blank {
		logging.WithField("error", "code cannot be blank").Fatal("Failed to call ReplyTo")

		return
	}

	// json-encoding the message
	jsonMessage, err := json.Marshal(m)
	if err != nil {
		logging.WithField("error", err.Error()).Fatal("Failed to call ReplyTo")

		return
	}

	if m.Code != codes.Ok {
		logging.WithFields(logrus.Fields{
			"error":          m.Err,
			"code":           m.Code,
			"reply_to":       natsMsg.Reply,
			"payload_length": len(jsonMessage),
		}).Error("Publishing an erroneous reply")
	} else {
		logging.WithFields(logrus.Fields{
			"reply_to":       natsMsg.Reply,
			"payload_length": len(jsonMessage),
			"code":           m.Code,
		}).Debug("Publishing a reply")
	}

	// attempting to Publish it
	err = mess.conn.Publish(natsMsg.Reply, jsonMessage)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":   err.Error(),
			"subject": natsMsg.Reply,
		}).Error("Failed to Publish message")

		logging.WithField("error", err.Error()).Fatal("Failed to call ReplyTo")

		return
	}
}

func (mess Messenger) Request(subject string, data []byte) (Message, error) {
	natsMsg, err := mess.conn.Request(subject, data, 5*time.Second)
	if err != nil {
		return Message{}, err
	}

	// json-decoding the message
	msg := &Message{}
	if err = json.Unmarshal(natsMsg.Data, &msg); err != nil {
		return Message{}, err
	}

	return *msg, nil
}

func (mess Messenger) Publish(subject string, data []byte) error {
	return mess.conn.Publish(subject, data)
}
