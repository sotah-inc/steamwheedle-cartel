package prod

import (
	"git.sotah.info/steamwheedle-cartel/pkg/bus"
	"git.sotah.info/steamwheedle-cartel/pkg/database"
	"git.sotah.info/steamwheedle-cartel/pkg/hell"
	"git.sotah.info/steamwheedle-cartel/pkg/logging"
	"git.sotah.info/steamwheedle-cartel/pkg/messenger"
	"git.sotah.info/steamwheedle-cartel/pkg/metric"
	"git.sotah.info/steamwheedle-cartel/pkg/state"
	"git.sotah.info/steamwheedle-cartel/pkg/state/subjects"
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

	// establishing bus-listeners
	sta.BusListeners = state.NewBusListeners(state.SubjectBusListeners{
		subjects.SyncPubsubTopicsMonitor: sta.ListenForSyncPubsubTopicsMonitor,
	})

	return sta, nil
}

type PubsubTopicsMonitorState struct {
	state.State

	actEndpoints hell.ActEndpoints
}
