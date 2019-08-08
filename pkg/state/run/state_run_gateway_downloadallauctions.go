package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	bCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/gameversions"
)

func (sta GatewayState) PublishDownloadedRegionRealmTuples(tuples sotah.RegionRealmTimestampTuples) error {
	// gathering a whitelist of region-realm-slugs
	regionRealmSlugs := tuples.ToRegionRealmSlugs()

	// gathering hell-realms for syncing
	logging.Info("Fetching region-realms from hell")
	hellRegionRealms, err := sta.IO.HellClient.GetRegionRealms(regionRealmSlugs, gameversions.Retail)
	if err != nil {
		return err
	}

	// updating the list of realms' timestamps
	logging.WithField(
		"total",
		hellRegionRealms.Total(),
	).Info("Updating region-realms in hell with new downloaded timestamp")
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

	// publishing region-realm slugs to the receive-realms messenger endpoint
	jsonEncoded, err := json.Marshal(regionRealmSlugs)
	if err != nil {
		return err
	}

	logging.Info("Publishing to receive-realms bus endpoint")
	req, err := sta.IO.BusClient.Request(sta.receiveRealmsTopic, string(jsonEncoded), 10*time.Second)
	if err != nil {
		return err
	}

	if req.Code != bCodes.Ok {
		return errors.New(req.Err)
	}

	return nil
}

func (sta GatewayState) PublishToCallComputeAllLiveAuctions(tuples sotah.RegionRealmTimestampTuples) error {
	// encoding the tuples for processing
	jsonEncoded, err := tuples.EncodeForDelivery()
	if err != nil {
		return err
	}

	logging.Info("Publishing to call-compute-all-live-auctions bus endpoint")
	req, err := sta.IO.BusClient.Request(sta.callComputeAllLiveAuctionsTopic, jsonEncoded, 10*time.Second)
	if err != nil {
		return err
	}

	if req.Code != bCodes.Ok {
		return errors.New(req.Err)
	}

	return nil
}

func (sta GatewayState) PublishToCallComputeAllPricelistHistories(tuples sotah.RegionRealmTimestampTuples) error {
	// encoding the tuples for processing
	jsonEncoded, err := tuples.EncodeForDelivery()
	if err != nil {
		return err
	}

	logging.Info("Publishing to call-compute-all-pricelist-histories bus endpoint")
	req, err := sta.IO.BusClient.Request(sta.callComputeAllPricelistHistoriesTopic, jsonEncoded, 10*time.Second)
	if err != nil {
		return err
	}

	if req.Code != bCodes.Ok {
		return errors.New(req.Err)
	}

	return nil
}

func (sta GatewayState) DownloadRegionRealms(
	regionRealms sotah.RegionRealms,
) (sotah.RegionRealmTimestampTuples, error) {
	// generating new act client
	logging.WithField(
		"endpoint-url",
		sta.actEndpoints.DownloadAuctions,
	).Info("Producing act client for download-auctions act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.DownloadAuctions)
	if err != nil {
		return sotah.RegionRealmTimestampTuples{}, err
	}

	// calling act client with region-realms
	logging.Info("Calling download-auctions with act client")
	actStartTime := time.Now()
	tuples := sotah.RegionRealmTimestampTuples{}
	totalIngestedBytes := 0
	for outJob := range actClient.DownloadAuctions(regionRealms) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to fetch auctions")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusCreated:
			// parsing the response body
			tuple, err := sotah.NewRegionRealmTimestampSizeTuple(string(outJob.Data.Body))
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": outJob.RegionName,
					"realm":  outJob.RealmSlug,
				}).Error("Failed to decode region-realm-timestamp tuple from act response body")

				continue
			}

			tuples = append(tuples, tuple.RegionRealmTimestampTuple)
			totalIngestedBytes += tuple.SizeBytes
		case http.StatusNotModified:
			logging.WithFields(logrus.Fields{
				"region": outJob.RegionName,
				"realm":  outJob.RealmSlug,
			}).Info("Region-realm tuple was processed but no new auctions were found")
		default:
			logging.WithFields(logrus.Fields{
				"region":      outJob.RegionName,
				"realm":       outJob.RealmSlug,
				"status-code": outJob.Data.Code,
				"data":        fmt.Sprintf("%.50s", string(outJob.Data.Body)),
			}).Error("Response code for act call was invalid")
		}
	}

	// reporting duration to reporter
	durationInUs := int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000)
	logging.WithFields(logrus.Fields{
		"duration-in-ms":       durationInUs * 1000,
		"total-ingested-bytes": totalIngestedBytes,
	},
	).Info("Finished calling act download-auctions")

	// reporting metrics
	m := metric.Metrics{
		"download_all_auctions_duration":   int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000),
		"download_all_auctions_size_bytes": totalIngestedBytes,
		"included_realms_downloaded":       len(tuples),
		"included_realms_total":            regionRealms.TotalRealms(),
	}
	if err := sta.IO.BusClient.PublishMetrics(m); err != nil {
		return sotah.RegionRealmTimestampTuples{}, err
	}

	return tuples, nil
}

func (sta GatewayState) DownloadAllAuctions() error {
	// gathering regions from boot-bucket
	regions, err := sta.bootBase.GetRegions(sta.bootBucket)
	if err != nil {
		return err
	}

	// gathering realms for each region from the realms base
	regionRealms := sotah.RegionRealms{}
	for _, region := range regions {
		realms, err := sta.realmsBase.GetAllRealms(region.Name, sta.realmsBucket)
		if err != nil {
			return err
		}

		regionRealms[region.Name] = realms
	}

	// downloading from all region-realms
	tuples, err := sta.DownloadRegionRealms(regionRealms)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to download all region-realms")

		return err
	}

	// optionally halting on no results
	if len(tuples) == 0 {
		logging.Info("No realms were updated")

		return nil
	}

	// publishing to receive-realms
	logging.Info("Publishing tuples to receive-realms")
	if err := sta.PublishDownloadedRegionRealmTuples(tuples); err != nil {
		return err
	}

	// publishing to call-compute-all-live-auctions
	logging.Info("Publishing tuples to call-compute-all-live-auctions")
	if err := sta.PublishToCallComputeAllLiveAuctions(tuples); err != nil {
		return err
	}

	// publishing to call-compute-all-pricelist-histories
	logging.Info("Publishing tuples to call-compute-all-pricelist-histories")
	if err := sta.PublishToCallComputeAllPricelistHistories(tuples); err != nil {
		return err
	}

	logging.Info("Finished")

	return nil
}
