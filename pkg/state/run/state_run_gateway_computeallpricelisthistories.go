package run

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	bCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/database"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta GatewayState) PublishComputedPricelistHistories(tuples sotah.RegionRealmTimestampTuples) error {
	requests := database.PricelistHistoriesComputeIntakeRequests{}
	for _, tuple := range tuples {
		requests = append(requests, database.PricelistHistoriesComputeIntakeRequest{
			RegionName:                tuple.RegionName,
			RealmSlug:                 tuple.RealmSlug,
			NormalizedTargetTimestamp: tuple.TargetTimestamp,
		})
	}

	// publishing to receive-computed-pricelist-histories bus endpoint
	encodedRequests, err := requests.EncodeForDelivery()
	if err != nil {
		return err
	}

	logging.Info("Publishing to receive-computed-pricelist-histories bus endpoint")
	req, err := sta.IO.BusClient.Request(sta.receiveComputedPricelistHistoriesTopic, encodedRequests, 10*time.Second)
	if err != nil {
		return err
	}

	if req.Code != bCodes.Ok {
		return errors.New(req.Err)
	}

	logging.Info("Finished pushing to receive-computed-pricelist-histories bus endpoint")

	return nil
}

func (sta GatewayState) ComputePricelistHistoriesFromTuples(
	tuples sotah.RegionRealmTimestampTuples,
) (sotah.RegionRealmTimestampTuples, error) {
	// generating new act client
	logging.WithField(
		"endpoint-url",
		sta.actEndpoints.ComputePricelistHistories,
	).Info("Producing act client for compute-pricelist-histories act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.ComputePricelistHistories)
	if err != nil {
		return sotah.RegionRealmTimestampTuples{}, err
	}

	// calling act client with region-realms
	logging.Info("Calling compute-pricelist-histories with act client")
	nextTuples := sotah.RegionRealmTimestampTuples{}
	actStartTime := time.Now()
	for outJob := range actClient.ComputePricelistHistories(tuples) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to compute pricelist-histories")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusCreated:
			// parsing the response body
			tuple, err := sotah.NewRegionRealmTimestampTuple(string(outJob.Data.Body))
			if err != nil {
				logging.WithFields(logrus.Fields{
					"error":  err.Error(),
					"region": outJob.RegionName,
					"realm":  outJob.RealmSlug,
				}).Error("Failed to decode region-realm-timestamp tuple from act response body")

				continue
			}

			nextTuples = append(nextTuples, tuple)
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
	durationInMs := int(int64(time.Since(actStartTime)) / 1000 / 1000)
	logging.WithFields(logrus.Fields{
		"duration-in-ms": durationInMs,
	},
	).Info("Finished calling act compute-pricelist-histories")

	// reporting metrics
	m := metric.Metrics{
		"compute_all_pricelist_histories_duration":     durationInMs / 1000,
		"included_realms_computed_pricelist_histories": len(tuples),
	}
	if err := sta.IO.BusClient.PublishMetrics(m); err != nil {
		return sotah.RegionRealmTimestampTuples{}, err
	}

	return nextTuples, nil
}

func (sta GatewayState) ComputeAllPricelistHistories(tuples sotah.RegionRealmTimestampTuples) error {
	// computing pricelist-histories from all region-realms
	nextTuples, err := sta.ComputePricelistHistoriesFromTuples(tuples)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to compute pricelist-histories from all region-realms")

		return err
	}

	// optionally halting on no results
	if len(tuples) == 0 {
		logging.Info("No realms were updated")

		return nil
	}

	// publishing to receive-realms
	logging.Info("Publishing region-realm tuples to receiver")
	if err := sta.PublishComputedPricelistHistories(nextTuples); err != nil {
		return err
	}

	logging.Info("Finished")

	return nil
}
