package stackdriver

import (
	"errors"
	"fmt"
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

var lsMap = levelSeverityMap{
	logrus.DebugLevel: syslog.LOG_DEBUG,
}

func (h Hook) Fire(entry *logrus.Entry) error {
	severity, ok := lsMap[entry.Level]
	if !ok {
		return errors.New("failed to resolve severity from level")
	}

	msg := &rfc5424.SyslogMessage{}
	msg.SetVersion(1)
	msg.SetMessage(entry.Message)
	msg.SetTimestamp(entry.Time.Format(time.RFC3339))
	msg.SetPriority(uint8(syslog.LOG_DAEMON | severity))

	if !msg.Valid() {
		return errors.New("rfc5424 message was not valid")
	}

	hookMessageBody, err := msg.String()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(h.conn, hookMessageBody); err != nil {
		return err
	}

	return nil
}

func (h Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
