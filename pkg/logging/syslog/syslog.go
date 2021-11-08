package stackdriver

import (
	"errors"
	"log/syslog"
	"net"
	"time"

	"github.com/influxdata/go-syslog/v3/rfc5424"
	"github.com/sirupsen/logrus"
)

func NewHook(protocol string, address string) (Hook, error) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		return Hook{}, err
	}

	return Hook{conn: conn}, nil
}

type Hook struct {
	conn net.Conn
}

type levelSeverityMap map[logrus.Level]syslog.Priority

const levelSeverityMap

func (h Hook) Fire(entry *logrus.Entry) error {
	msg := &rfc5424.SyslogMessage{}
	msg.SetVersion(1)
	msg.SetMessage(entry.Message)
	msg.SetTimestamp(entry.Time.Format(time.RFC3339))
	msg.SetPriority(uint8(syslog.LOG_DAEMON))

	if !msg.Valid() {
		return errors.New("rfc5424 message was not valid")
	}

	hookMessageBody, err := msg.String()
	if err != nil {
		return err
	}

	switch entry.Level {
	case logrus.DebugLevel:
		return h.writer.Debug(hookMessageBody)
	case logrus.InfoLevel:
		return h.writer.Info(hookMessageBody)
	}

	return errors.New("invalid entry level provided")
}

func (h Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
