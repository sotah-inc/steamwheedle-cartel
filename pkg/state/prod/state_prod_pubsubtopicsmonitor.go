package prod

import (
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/hell"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/messenger"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state"
	"github.com/twinj/uuid"
)

type PubsubTopicsMonitorStateConfig struct {
	ProjectID string

	MessengerHost string
	MessengerPort int

	PubsubTopicsDatabaseDir string
}

func NewPubsubTopicsMonitorState(config PubsubTopicsMonitorStateConfig) (PubsubTopicsMonitorState, error) {
	// establishing an initial state
	sta := PubsubTopicsMonitorState{
		State: state.NewState(uuid.NewV4(), true),
	}

	// connecting to the messenger host
	mess, err := messenger.NewMessenger(config.MessengerHost, config.MessengerPort)
	if err != nil {
		return PubsubTopicsMonitorState{}, err
	}
	sta.IO.Messenger = mess

	// initializing a reporter
	sta.IO.Reporter = metric.NewReporter(mess)

	// loading the items database
	logging.Info("Connecting to pubsub-topics database")
	pubsubTopicsDatabase, err := database.NewPubsubTopicsDatabase(config.PubsubTopicsDatabaseDir)
	if err != nil {
		return PubsubTopicsMonitorState{}, err
	}
	sta.IO.Databases.PubsubTopicsDatabase = pubsubTopicsDatabase

	// establishing a bus
	logging.Info("Connecting bus-client")
	busClient, err := bus.NewClient(config.ProjectID, "prod-pubsub-topics-monitor")
	if err != nil {
		return PubsubTopicsMonitorState{}, err
	}
	sta.IO.BusClient = busClient

	// connecting to hell
	sta.IO.HellClient, err = hell.NewClient(config.ProjectID)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to connect to hell")

		return PubsubTopicsMonitorState{}, err
	}

	// gathering act-endpoints
	sta.actEndpoints, err = sta.IO.HellClient.GetActEndpoints()
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to fetch act endpoints")

		return PubsubTopicsMonitorState{}, err
	}

	return sta, nil
}

type PubsubTopicsMonitorState struct {
	state.State

	actEndpoints hell.ActEndpoints
}
