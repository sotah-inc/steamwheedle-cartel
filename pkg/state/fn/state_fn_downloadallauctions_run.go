package fn

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/bus"
	bCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	mCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/messenger/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/state/subjects"
)

func (sta DownloadAllAuctionsState) PublishToReceiveRealms(tuples bus.RegionRealmTimestampTuples) error {
	// gathering a whitelist of region-realm-slugs
	regionRealmSlugs := tuples.ToRegionRealmSlugs()

	// gathering hell-realms for syncing
	hellRegionRealms, err := sta.IO.HellClient.GetRegionRealms(regionRealmSlugs, gameversions.Retail)
	if err != nil {
		return err
	}

	// updating the list of realms' timestamps
	for _, tuple := range tuples {
		hellRealm := hellRegionRealms[blizzard.RegionName(tuple.RegionName)][blizzard.RealmSlug(tuple.RealmSlug)]
		hellRealm.Downloaded = tuple.TargetTimestamp
		hellRegionRealms[blizzard.RegionName(tuple.RegionName)][blizzard.RealmSlug(tuple.RealmSlug)] = hellRealm

		logrus.WithFields(logrus.Fields{
			"region":     blizzard.RegionName(tuple.RegionName),
			"realm":      blizzard.RealmSlug(tuple.RealmSlug),
			"downloaded": tuple.TargetTimestamp,
		}).Info("Setting downloaded value for hell realm")
	}
	if err := sta.IO.HellClient.WriteRegionRealms(hellRegionRealms, gameversions.Retail); err != nil {
		return err
	}

	jsonEncoded, err := json.Marshal(regionRealmSlugs)
	if err != nil {
		return err
	}

	req, err := sta.IO.Messenger.Request(string(subjects.ReceiveRealms), jsonEncoded)
	if err != nil {
		return err
	}

	if req.Code != mCodes.Ok {
		return errors.New(req.Err)
	}

	return nil
}

func (sta DownloadAllAuctionsState) Run() error {
	regions, err := sta.bootBase.GetRegions(sta.bootBucket)
	if err != nil {
		return err
	}

	regionRealms := sotah.RegionRealms{}
	for _, region := range regions {
		realms, err := sta.realmsBase.GetAllRealms(region.Name, sta.realmsBucket)
		if err != nil {
			return err
		}

		regionRealms[region.Name] = realms
	}

	logging.WithField("endpoint-url", sta.actEndpoints.DownloadAuctions).Info("Producing act client")
	actClient, err := act.NewClient(sta.actEndpoints.DownloadAuctions)
	if err != nil {
		return err
	}

	logging.Info("Calling download-auctions with act client")
	actStartTime := time.Now()
	for outJob := range actClient.DownloadAuctions(regionRealms) {
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to fetch auctions")

			continue
		}

		logging.WithFields(logrus.Fields{
			"region": outJob.RegionName,
			"realm":  outJob.RealmSlug,
			"data":   string(outJob.Data),
		}).Info("Received from download-auctions")
	}
	logging.WithField(
		"duration",
		int(int64(time.Since(actStartTime))/1000/1000/1000),
	).Info("Finished calling act download-auctions")

	// producing messages
	logging.Info("Producing messages for bulk requesting")
	messages, err := bus.NewCollectAuctionMessages(regionRealms)
	if err != nil {
		return err
	}

	// enqueueing them and gathering result jobs
	startTime := time.Now()
	responseItems, err := sta.IO.BusClient.BulkRequest(sta.downloadAuctionsTopic, messages, 120*time.Second)
	if err != nil {
		return err
	}

	// gathering validated response messages
	validatedResponseItems := bus.BulkRequestMessages{}
	for k, msg := range responseItems {
		if msg.Code != bCodes.Ok {
			if msg.Code == bCodes.BlizzardError {
				var respError blizzard.ResponseError
				if err := json.Unmarshal([]byte(msg.Data), &respError); err != nil {
					return err
				}

				logging.WithFields(logrus.Fields{"resp-error": respError}).Error("Received erroneous response")
			}

			continue
		}

		// ok msg code but no msg data means no new auctions
		if len(msg.Data) == 0 {
			continue
		}

		validatedResponseItems[k] = msg
	}

	// reporting metrics
	if err := sta.IO.BusClient.PublishMetrics(metric.Metrics{
		"download_all_auctions_duration": int(int64(time.Since(startTime)) / 1000 / 1000 / 1000),
		"included_realms":                len(validatedResponseItems),
	}); err != nil {
		return err
	}

	// formatting the response-items as tuples for processing
	tuples, err := bus.NewRegionRealmTimestampTuplesFromMessages(validatedResponseItems)
	if err != nil {
		return err
	}

	// publishing to receive-realms
	logging.Info("Publishing realms to receive-realms")
	if err := sta.PublishToReceiveRealms(tuples); err != nil {
		return err
	}

	// encoding tuples for publishing to compute-all-live-auctions and compute-all-pricelist-histories
	encodedTuples, err := tuples.EncodeForDelivery()
	if err != nil {
		return err
	}
	encodedTuplesMsg := bus.NewMessage()
	encodedTuplesMsg.Data = encodedTuples

	// publishing to compute-all-live-auctions
	logging.WithField("tuples", len(tuples)).Info("Publishing to compute-all-live-auctions")
	if _, err := sta.IO.BusClient.Publish(sta.computeAllLiveAuctionsTopic, encodedTuplesMsg); err != nil {
		return err
	}

	// publishing to compute-all-pricelist-histories
	logging.Info("Publishing to compute-all-live-auctions")
	if _, err := sta.IO.BusClient.Publish(sta.computeAllPricelistHistoriesTopic, encodedTuplesMsg); err != nil {
		return err
	}

	logging.Info("Finished")

	return nil
}
