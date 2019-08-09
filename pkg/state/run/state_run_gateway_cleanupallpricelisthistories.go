package run

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta GatewayState) CleanupRegionRealmsPricelistHistories(regionRealms sotah.RegionRealms) error {
	// generating new act client
	logging.WithField(
		"endpoint-url",
		sta.actEndpoints.CleanupPricelistHistories,
	).Info("Producing act client for cleanup-pricelist-histories act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.CleanupPricelistHistories)
	if err != nil {
		return err
	}

	// calling act client with region-realms
	logging.Info("Calling cleanup-pricelist-histories with act client")
	actStartTime := time.Now()
	totalDeleted := 0
	for outJob := range actClient.CleanupPricelistHistories(regionRealms) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to cleanup pricelist-histories")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusOK:
			// parsing the response body
			resp, err := sotah.NewCleanupPricelistPayloadResponse(string(outJob.Data.Body))
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": outJob.RegionName,
					"realm":  outJob.RealmSlug,
				}).Error("Failed to decode cleanup-pricelist-histories-payload-response from act response body")

				continue
			}

			totalDeleted += resp.TotalDeleted

			logging.WithFields(logrus.Fields{
				"region":        resp.RegionName,
				"realm":         resp.RealmSlug,
				"total-deleted": resp.TotalDeleted,
			}).Info("Pricelist-histories have been cleaned")
		default:
			logging.WithFields(logrus.Fields{
				"region":      outJob.RegionName,
				"realm":       outJob.RealmSlug,
				"status-code": outJob.Data.Code,
				"data":        fmt.Sprintf("%.25s", string(outJob.Data.Body)),
			}).Error("Response code for act call was invalid")
		}
	}

	// reporting duration to reporter
	durationInUs := int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000)
	logging.WithFields(logrus.Fields{
		"duration-in-ms": durationInUs * 1000,
		"total-deleted":  totalDeleted,
	}).Info("Finished calling act cleanup-pricelist-histories")

	// reporting metrics
	m := metric.Metrics{
		"cleanup_all_pricelist_histories_duration":      int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000),
		"cleanup_all_pricelist_histories_total_deleted": totalDeleted,
	}
	if err := sta.IO.BusClient.PublishMetrics(m); err != nil {
		return err
	}

	return nil
}

func (sta GatewayState) CleanupAllPricelistHistories() error {
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

	// cleaning up pricelist-histories in all region-realms
	if err := sta.CleanupRegionRealmsPricelistHistories(regionRealms); err != nil {
		logging.WithField("error", err.Error()).Error("Failed to cleanup all region-realms pricelist-histories")

		return err
	}

	logging.Info("Finished")

	return nil
}
