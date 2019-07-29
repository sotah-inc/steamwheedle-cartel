package prod

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

func HandleComputedLiveAuctions(liveAuctionsState ProdLiveAuctionsState, tuples sotah.RegionRealmTuples) {
	// declaring a load-in channel for the live-auctions db and starting it up
	loadInJobs := make(chan database.LiveAuctionsLoadEncodedDataInJob)
	loadOutJobs := liveAuctionsState.IO.Databases.LiveAuctionsDatabases.LoadEncodedData(loadInJobs)

	// starting workers for handling tuples
	in := make(chan sotah.RegionRealmTuple)
	worker := func() {
		for tuple := range in {
			// resolving the realm from the request
			realm, err := func() (sotah.Realm, error) {
				for regionName, status := range liveAuctionsState.Statuses {
					if regionName != blizzard.RegionName(tuple.RegionName) {
						continue
					}

					for _, realm := range status.Realms {
						if realm.Slug != blizzard.RealmSlug(tuple.RealmSlug) {
							continue
						}

						return realm, nil
					}
				}

				return sotah.Realm{}, errors.New("realm not found")
			}()
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to resolve realm from tuple")

				continue
			}

			// resolving the data
			data, err := func() ([]byte, error) {
				obj, err := liveAuctionsState.LiveAuctionsBase.GetFirmObject(realm, liveAuctionsState.LiveAuctionsBucket)
				if err != nil {
					return []byte{}, err
				}

				reader, err := obj.ReadCompressed(true).NewReader(liveAuctionsState.IO.StoreClient.Context)
				if err != nil {
					return []byte{}, err
				}

				out, err := ioutil.ReadAll(reader)
				if err != nil {
					return []byte{}, err
				}
				if err := reader.Close(); err != nil {
					return []byte{}, err
				}

				return out, nil
			}()
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to get data")

				continue
			}

			loadInJobs <- database.LiveAuctionsLoadEncodedDataInJob{
				RegionName:  blizzard.RegionName(tuple.RegionName),
				RealmSlug:   blizzard.RealmSlug(tuple.RealmSlug),
				EncodedData: data,
			}
		}
	}
	postWork := func() {
		close(loadInJobs)
	}
	util.Work(4, worker, postWork)

	// queueing it all up
	go func() {
		for _, tuple := range tuples {
			logging.WithFields(logrus.Fields{
				"region": tuple.RegionName,
				"realm":  tuple.RealmSlug,
			}).Info("Loading tuple")

			in <- tuple
		}

		close(in)
	}()

	// waiting for the results to drain out
	for job := range loadOutJobs {
		if job.Err != nil {
			logging.WithFields(job.ToLogrusFields()).Error("Failed to load job")

			continue
		}

		logging.WithFields(logrus.Fields{
			"region": job.RegionName,
			"realm":  job.RealmSlug,
		}).Info("Loaded job")
	}
}

func (liveAuctionsState ProdLiveAuctionsState) ListenForComputedLiveAuctions(
	onReady chan interface{},
	stop chan interface{},
	onStopped chan interface{},
) {
	// establishing subscriber config
	config := bus.SubscribeConfig{
		Stop: stop,
		Callback: func(busMsg bus.Message) {
			// decoding message body
			tuples, err := sotah.NewRegionRealmTuples(busMsg.Data)
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to decode region-realm tuples")

				if err := liveAuctionsState.IO.BusClient.ReplyToWithError(busMsg, err, codes.GenericError); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply to message")

					return
				}

				return
			}

			// acking the message
			if _, err := liveAuctionsState.IO.BusClient.ReplyTo(busMsg, bus.NewMessage()); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to reply to message")

				return
			}

			// handling requests
			logging.WithField("requests", len(tuples)).Info("Received tuples")
			startTime := time.Now()
			HandleComputedLiveAuctions(liveAuctionsState, tuples)
			logging.WithField("requests", len(tuples)).Info("Done handling tuples")

			// reporting metrics
			m := metric.Metrics{
				"receive_all_live_auctions_duration": int(int64(time.Since(startTime)) / 1000 / 1000 / 1000),
			}
			if err := liveAuctionsState.IO.BusClient.PublishMetrics(m); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to publish metric")

				if err := liveAuctionsState.IO.BusClient.ReplyToWithError(busMsg, err, codes.GenericError); err != nil {
					logging.WithField("error", err.Error()).Error("Failed to reply to message")

					return
				}

				return
			}

			// gathering hell-realms for syncing
			logging.Info("Fetching region-realms from hell")
			hellRegionRealms, err := liveAuctionsState.IO.HellClient.GetRegionRealms(
				tuples.ToRegionRealmSlugs(),
				gameversions.Retail,
			)
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to get region-realms")

				return
			}

			// updating the list of realms' timestamps
			logging.WithField(
				"total",
				hellRegionRealms.Total(),
			).Info("Updating region-realms in hell with new downloaded timestamp")
			for _, tuple := range tuples {
				hellRealm := hellRegionRealms[blizzard.RegionName(tuple.RegionName)][blizzard.RealmSlug(tuple.RealmSlug)]
				hellRealm.LiveAuctionsReceived = int(time.Now().Unix())
				hellRegionRealms[blizzard.RegionName(tuple.RegionName)][blizzard.RealmSlug(tuple.RealmSlug)] = hellRealm

				logrus.WithFields(logrus.Fields{
					"region":     blizzard.RegionName(tuple.RegionName),
					"realm":      blizzard.RealmSlug(tuple.RealmSlug),
					"downloaded": hellRealm.LiveAuctionsReceived,
				}).Info("Setting downloaded value for hell realm")
			}
			if err := liveAuctionsState.IO.HellClient.WriteRegionRealms(hellRegionRealms, gameversions.Retail); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to write region-realms to hell")

				return
			}

			// publishing region-realm slugs to the receive-realms messenger endpoint
			jsonEncoded, err := json.Marshal(tuples.ToRegionRealmSlugs())
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to encode region-realm slugs for publishing")

				return
			}

			logging.Info("Publishing to receive-realms bus endpoint")
			req, err := liveAuctionsState.IO.BusClient.Request(
				liveAuctionsState.receiveRealmsTopic,
				string(jsonEncoded),
				10*time.Second,
			)
			if err != nil {
				logging.WithField("error", err.Error()).Error("Failed to encode region-realm slugs for publishing")

				return
			}

			if req.Code != codes.Ok {
				logging.WithField(
					"error",
					errors.New("response code was not ok").Error(),
				).Error("Publish succeeded but response code was not ok")

				return
			}
		},
		OnReady:   onReady,
		OnStopped: onStopped,
	}

	// starting up worker for the subscription
	go func() {
		if err := liveAuctionsState.IO.BusClient.SubscribeToTopic(
			string(subjects.ReceiveComputedLiveAuctions),
			config,
		); err != nil {
			logging.WithField("error", err.Error()).Fatal("Failed to subscribe to topic")
		}
	}()
}
