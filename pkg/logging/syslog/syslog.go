package syslog

import (
	"errors"
	"fmt"
	"log/syslog"
	"net"
	"strings"
	"time"

	"github.com/influxdata/go-syslog/v3/rfc5424"
	"github.com/sirupsen/logrus"
)

func NewHook(network string, address string) (Hook, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return Hook{}, err
	}

	h := Hook{
		network: network,
		address: address,
		conn:    conn,
		connIn:  make(chan string),
	}

	go func() {
		for msg := range h.connIn {
			if err := h.send(msg, 0); err != nil {
				fmt.Printf("failed to send message: %s\n", err.Error())

				continue
			}
		}
	}()

	return h, nil
}

type Hook struct {
	network string
	address string
	conn    net.Conn
	connIn  chan string
}

const ErrBrokenPipe = "write: broken pipe"

func (h Hook) send(msg string, attempt int) error {
	if attempt > 5 {
		return errors.New("maximum send attempt reached")
	}

	if _, err := fmt.Fprint(h.conn, msg); err != nil {
		//fmt.Printf("received error: %s\n", err.Error())

		if !strings.HasSuffix(err.Error(), ErrBrokenPipe) {
			//fmt.Printf("error was not %s: %s\n", ErrBrokenPipe, err.Error())

			return err
		}

		//fmt.Println("reconnecting")

		conn, err := net.Dial(h.network, h.address)
		if err != nil {
			//fmt.Printf("failed to reconnect: %s, attempting to reconnect\n", err.Error())

			return h.send(msg, attempt+1)
		}

		h.conn = conn

		//fmt.Printf("resending on attempt: %d\n", attempt+1)

		return h.send(msg, attempt+1)
	}

	return nil
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
	for field, value := range entry.Data {
		parsedValue, ok := value.(string)
		if !ok {
			continue
		}

		msg.SetParameter("examplesdid@32473", field, parsedValue)
	}

	if !msg.Valid() {
		return errors.New("rfc5424 message was not valid")
	}

	hookMessageBody, err := msg.String()
	if err != nil {
		return err
	}

	fullHookMessageBody := fmt.Sprintf("%d %s\n", len(hookMessageBody)+1, hookMessageBody)

	//fmt.Printf("sending log line: %s\n", fullHookMessageBody)

	h.connIn <- fullHookMessageBody

	return nil
}

func (h Hook) Levels() []logrus.Level {
	return logrus.AllLevels
}
