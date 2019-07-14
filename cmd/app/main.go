package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel-server/app/commands"
	devCommand "github.com/sotah-inc/steamwheedle-cartel/pkg/command/dev"
	fnCommand "github.com/sotah-inc/steamwheedle-cartel/pkg/command/fn"
	prodCommand "github.com/sotah-inc/steamwheedle-cartel/pkg/command/prod"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging/stackdriver"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	devState "github.com/sotah-inc/steamwheedle-cartel/pkg/state/dev"
	fnState "github.com/sotah-inc/steamwheedle-cartel/pkg/state/fn"
	prodState "github.com/sotah-inc/steamwheedle-cartel/pkg/state/prod"
	"github.com/twinj/uuid"
	"gopkg.in/alecthomas/kingpin.v2"
)

type commandMap map[string]func() error

// ID represents this run's unique id
var ID uuid.UUID

func main() {
	// assigning global ID
	ID = uuid.NewV4()

	// parsing the command flags
	var (
		app            = kingpin.New("sotah-server", "A command-line Blizzard AH client.")
		natsHost       = app.Flag("nats-host", "NATS hostname").Default("localhost").Envar("NATS_HOST").Short('h').String()
		natsPort       = app.Flag("nats-port", "NATS port").Default("4222").Envar("NATS_PORT").Short('p').Int()
		configFilepath = app.Flag("config", "Relative path to config json").Required().Short('c').String()
		clientID       = app.Flag("client-id", "Blizzard API Client ID").Envar("CLIENT_ID").String()
		clientSecret   = app.Flag("client-secret", "Blizzard API Client Secret").Envar("CLIENT_SECRET").String()
		verbosity      = app.Flag("verbosity", "Log verbosity").Default("info").Short('v').String()
		cacheDir       = app.Flag("cache-dir", "Directory to cache data files to").Required().String()
		projectID      = app.Flag("project-id", "GCloud Storage Project ID").Default("").Envar("PROJECT_ID").String()

		apiCommand                = app.Command(string(commands.API), "For running sotah-server.")
		liveAuctionsCommand       = app.Command(string(commands.LiveAuctions), "For in-memory storage of current auctions.")
		pricelistHistoriesCommand = app.Command(string(commands.PricelistHistories), "For on-disk storage of pricelist histories.")

		prodApiCommand                = app.Command(string(commands.ProdApi), "For running sotah-server in prod-mode.")
		prodMetricsCommand            = app.Command(string(commands.ProdMetrics), "For forwarding metrics to a nats channel.")
		prodLiveAuctionsCommand       = app.Command(string(commands.ProdLiveAuctions), "For managing live-auctions in gcp ce vm.")
		prodPricelistHistoriesCommand = app.Command(string(commands.ProdPricelistHistories), "For managing pricelist-histories in gcp ce vm.")
		prodItemsCommand              = app.Command(string(commands.ProdItems), "For managing items in gcp ce vm.")

		fnDownloadAllAuctions          = app.Command(string(commands.FnDownloadAllAuctions), "For enqueueing downloading of auctions in gcp ce vm.")
		fnComputeAllLiveAuctions       = app.Command(string(commands.FnComputeAllLiveAuctions), "For enqueueing computing of all live-auctions in gcp ce vm.")
		fnComputeAllPricelistHistories = app.Command(string(commands.FnComputeAllPricelistHistories), "For enqueueing computing of all live-auctions in gcp ce vm.")
		fnSyncAllItems                 = app.Command(string(commands.FnSyncAllItems), "For enqueueing syncing of items and item-icons in gcp ce vm.")
		fnCleanupAllExpiredManifests   = app.Command(string(commands.FnCleanupAllExpiredManifests), "For gathering all expired auction-manifests for deletion in gcp ce vm.")
		fnCleanupPricelistHistories    = app.Command(string(commands.FnCleanupPricelistHistories), "For gathering all expired pricelist-histories for deletion in gcp ce vm.")
	)
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	// establishing log verbosity
	logVerbosity, err := logrus.ParseLevel(*verbosity)
	if err != nil {
		logging.WithField("error", err.Error()).Fatal("Could not parse log level")

		return
	}
	logging.SetLevel(logVerbosity)

	// loading the config file
	c, err := sotah.NewConfigFromFilepath(*configFilepath)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error":    err.Error(),
			"filepath": *configFilepath,
		}).Fatal("Could not fetch config")

		return
	}

	// optionally adding stackdriver hook
	if c.UseGCloud {
		logging.WithField("project-id", *projectID).Info("Creating stackdriver hook")
		stackdriverHook, err := stackdriver.NewHook(*projectID, cmd)
		if err != nil {
			logging.WithFields(logrus.Fields{
				"error":     err.Error(),
				"projectID": projectID,
			}).Fatal("Could not create new stackdriver logrus hook")

			return
		}

		logging.AddHook(stackdriverHook)
	}
	logging.Info("Starting")

	logging.WithField("command", cmd).Info("Running command")

	// declaring a command map
	cMap := commandMap{
		apiCommand.FullCommand(): func() error {
			return devCommand.Api(devState.APIStateConfig{
				SotahConfig:          c,
				DiskStoreCacheDir:    *cacheDir,
				ItemsDatabaseDir:     fmt.Sprintf("%s/databases", *cacheDir),
				BlizzardClientSecret: *clientSecret,
				BlizzardClientId:     *clientID,
				MessengerPort:        *natsPort,
				MessengerHost:        *natsHost,
				GCloudProjectID:      *projectID,
			})
		},
		liveAuctionsCommand.FullCommand(): func() error {
			return devCommand.LiveAuctions(devState.LiveAuctionsStateConfig{
				MessengerHost:           *natsHost,
				MessengerPort:           *natsPort,
				DiskStoreCacheDir:       *cacheDir,
				LiveAuctionsDatabaseDir: fmt.Sprintf("%s/databases", *cacheDir),
			})
		},
		pricelistHistoriesCommand.FullCommand(): func() error {
			return devCommand.PricelistHistories(devState.PricelistHistoriesStateConfig{
				DiskStoreCacheDir:             *cacheDir,
				MessengerPort:                 *natsPort,
				MessengerHost:                 *natsHost,
				PricelistHistoriesDatabaseDir: fmt.Sprintf("%s/databases", *cacheDir),
			})
		},
		prodApiCommand.FullCommand(): func() error {
			return prodCommand.ProdApi(prodState.ProdApiStateConfig{
				SotahConfig:     c,
				MessengerPort:   *natsPort,
				MessengerHost:   *natsHost,
				GCloudProjectID: *projectID,
			})
		},
		prodMetricsCommand.FullCommand(): func() error {
			return prodCommand.ProdMetrics(prodState.ProdMetricsStateConfig{
				MessengerPort:   *natsPort,
				MessengerHost:   *natsHost,
				GCloudProjectID: *projectID,
			})
		},
		prodLiveAuctionsCommand.FullCommand(): func() error {
			return prodCommand.ProdLiveAuctions(prodState.ProdLiveAuctionsStateConfig{
				MessengerPort:           *natsPort,
				MessengerHost:           *natsHost,
				GCloudProjectID:         *projectID,
				LiveAuctionsDatabaseDir: fmt.Sprintf("%s/databases", *cacheDir),
			})
		},
		prodPricelistHistoriesCommand.FullCommand(): func() error {
			return prodCommand.ProdPricelistHistories(prodState.ProdPricelistHistoriesStateConfig{
				MessengerPort:                 *natsPort,
				MessengerHost:                 *natsHost,
				GCloudProjectID:               *projectID,
				PricelistHistoriesDatabaseDir: fmt.Sprintf("%s/databases", *cacheDir),
			})
		},
		prodItemsCommand.FullCommand(): func() error {
			return prodCommand.ProdItems(prodState.ProdItemsStateConfig{
				MessengerPort:    *natsPort,
				MessengerHost:    *natsHost,
				GCloudProjectID:  *projectID,
				ItemsDatabaseDir: fmt.Sprintf("%s/databases", *cacheDir),
			})
		},
		fnDownloadAllAuctions.FullCommand(): func() error {
			return fnCommand.FnDownloadAllAuctions(fnState.DownloadAllAuctionsStateConfig{
				ProjectId:     *projectID,
				MessengerHost: *natsHost,
				MessengerPort: *natsPort,
			})
		},
		fnComputeAllLiveAuctions.FullCommand(): func() error {
			return fnCommand.FnComputeAllLiveAuctions(fnState.ComputeAllLiveAuctionsStateConfig{
				ProjectId: *projectID,
			})
		},
		fnComputeAllPricelistHistories.FullCommand(): func() error {
			return fnCommand.FnComputeAllPricelistHistories(fnState.ComputeAllPricelistHistoriesStateConfig{
				ProjectId: *projectID,
			})
		},
		fnSyncAllItems.FullCommand(): func() error {
			return fnCommand.FnSyncAllItems(fnState.SyncAllItemsStateConfig{
				ProjectId: *projectID,
			})
		},
		fnCleanupAllExpiredManifests.FullCommand(): func() error {
			return fnCommand.FnCleanupAllExpiredManifests(fnState.CleanupAllExpiredManifestsStateConfig{
				ProjectId: *projectID,
			})
		},
		fnCleanupPricelistHistories.FullCommand(): func() error {
			return fnCommand.FnCleanupPricelistHistories(fnState.CleanupPricelistHistoriesStateConfig{
				ProjectId: *projectID,
			})
		},
	}

	// resolving the command func
	cmdFunc, ok := cMap[cmd]
	if !ok {
		logging.WithField("command", cmd).Fatal("Invalid command")

		return
	}

	// calling the command func
	if err := cmdFunc(); err != nil {
		logging.WithFields(logrus.Fields{
			"error":   err.Error(),
			"command": cmd,
		}).Fatal("Failed to execute command")
	}
}
