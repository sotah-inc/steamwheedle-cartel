package metric

import (
	"encoding/json"
	"fmt"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/metric/kinds"
)

const (
	appMetricSubject = "appMetrics"
)

func NewReporter(mess messenger.Messenger) Reporter {
	return Reporter{mess}
}

type Reporter struct {
	Messenger messenger.Messenger
}

type Metrics map[string]int

func (re Reporter) Report(m Metrics) {
	data, err := json.Marshal(m)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to marshal report metric")

		return
	}

	if err := re.Messenger.Publish(appMetricSubject, data); err != nil {
		logging.WithField("error", err.Error()).Error("Failed to publish to app-metrics subject")

		return
	}

	return
}

func (re Reporter) ReportWithPrefix(m Metrics, prefix kinds.Kind) {
	next := Metrics{}

	for k, v := range m {
		next[fmt.Sprintf("%s_%s", prefix, k)] = v
	}

	re.Report(next)
}
