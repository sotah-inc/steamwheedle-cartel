package syslog

import (
	"errors"
	"fmt"
	"log/syslog"
	"net"
	"time"

	"github.com/influxdata/go-syslog/v3/rfc5424"
	"github.com/sirupsen/logrus"
)

func NewHook(network string, address string) (Hook, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return Hook{}, err
	}

	connIn := make(chan string)
	go func() {
		for msg := range connIn {
			if _, err := fmt.Fprint(conn, msg); err != nil {
				fmt.Printf("failed to write to syslog endpoint: %s\n", err.Error())
			}
		}
	}()

	return Hook{conn: conn, connIn: connIn}, nil
}

type Hook struct {
	conn   net.Conn
	connIn chan string
}

type levelSeverityMap map[logrus.Level]syslog.Priority

var lsMap = levelSeverityMap{
	logrus.DebugLevel: syslog.LOG_DEBUG,
	logrus.TraceLevel: syslog.LOG_DEBUG,
	logrus.InfoLevel:  syslog.LOG_INFO,
	logrus.WarnLevel:  syslog.LOG_WARNING,
	logrus.ErrorLevel: syslog.LOG_ERR,
	logrus.FatalLevel: syslog.LOG_CRIT,
	logrus.PanicLevel: syslog.LOG_CRIT,
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

	fullHookMessageBody := fmt.Sprintf("%d %s", len(hookMessageBody)+1, hookMessageBody)

	fmt.Printf("sending log line: %s\n", fullHookMessageBody)

	h.connIn <- fullHookMessageBody

	return nil
}

func (h Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
