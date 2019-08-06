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

func (sta GatewayState) CleanupRegionRealmsAuctions(regionRealms sotah.RegionRealms) error {
	// generating new act client
	logging.WithField(
		"endpoint-url",
		sta.actEndpoints.CleanupAuctions,
	).Info("Producing act client for cleanup-auctions act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.CleanupAuctions)
	if err != nil {
		return err
	}

	// calling act client with region-realms
	logging.Info("Calling cleanup-auctions with act client")
	actStartTime := time.Now()
	totalDeletedSizeBytes := int64(0)
	totalDeletedCount := 0
	for outJob := range actClient.CleanupAuctions(regionRealms) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to cleanup auctions")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusOK:
			// parsing the response body
			resp, err := sotah.NewCleanupAuctionsPayloadResponse(string(outJob.Data.Body))
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": outJob.RegionName,
					"realm":  outJob.RealmSlug,
				}).Error("Failed to decode cleanup-auctions-payload-response from act response body")

				continue
			}

			totalDeletedSizeBytes += resp.TotalDeletedSizeBytes
			totalDeletedCount += resp.TotalDeletedCount

			logging.WithFields(logrus.Fields{
				"region":                   resp.RegionName,
				"realm":                    resp.RealmSlug,
				"total-deleted-count":      resp.TotalDeletedCount,
				"total-deleted-size-bytes": resp.TotalDeletedSizeBytes,
			}).Info("Auctions have been cleaned")
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
		"duration-in-ms":           durationInUs * 1000,
		"total-deleted-count":      totalDeletedCount,
		"total-deleted-size-bytes": totalDeletedSizeBytes,
	}).Info("Finished calling act cleanup-auctions")

	// reporting metrics
	m := metric.Metrics{
		"cleanup_all_auctions_duration":            int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000),
		"cleanup_all_auctions_total_deleted":       totalDeletedCount,
		"cleanup_all_auctions_total_deleted_bytes": int(totalDeletedSizeBytes),
	}
	if err := sta.IO.BusClient.PublishMetrics(m); err != nil {
		return err
	}

	return nil
}

func (sta GatewayState) CleanupAllAuctions() error {
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

	// cleaning up auctions in all region-realms
	if err := sta.CleanupRegionRealmsAuctions(regionRealms); err != nil {
		logging.WithField("error", err.Error()).Error("Failed to cleanup all region-realms auctions")

		return err
	}

	logging.Info("Finished")

	return nil
}
