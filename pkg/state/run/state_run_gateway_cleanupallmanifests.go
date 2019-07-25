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

func (sta GatewayState) CleanupRegionRealms(regionRealms sotah.RegionRealms) error {
	// generating new act client
	logging.WithField("endpoint-url", sta.actEndpoints.CleanupManifests).Info("Producing act client")
	actClient, err := act.NewClient(sta.actEndpoints.CleanupManifests)
	if err != nil {
		return err
	}

	// calling act client with region-realms
	logging.Info("Calling cleanup-manifests with act client")
	actStartTime := time.Now()
	for outJob := range actClient.CleanupManifests(regionRealms) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to cleanup manifests")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusOK:
			// parsing the response body
			tuple, err := sotah.NewRegionRealmTuple(string(outJob.Data.Body))
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": outJob.RegionName,
					"realm":  outJob.RealmSlug,
				}).Error("Failed to decode region-realm tuple from act response body")

				continue
			}

			logging.WithFields(logrus.Fields{
				"region": tuple.RegionName,
				"realm":  tuple.RealmSlug,
			}).Info("Manifests have been cleaned")
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
	logging.WithField(
		"duration-in-ms",
		durationInUs*1000,
	).Info("Finished calling act cleanup-manifests")

	// reporting metrics
	m := metric.Metrics{
		"cleanup_all_manifests_duration": int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000),
	}
	if err := sta.IO.BusClient.PublishMetrics(m); err != nil {
		return err
	}

	return nil
}

func (sta GatewayState) CleanupAllManifests() error {
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

	// cleaning up manifests in all region-realms
	if err := sta.CleanupRegionRealms(regionRealms); err != nil {
		logging.WithField("error", err.Error()).Error("Failed to cleanup all region-realms")

		return err
	}

	logging.Info("Finished")

	return nil
}
