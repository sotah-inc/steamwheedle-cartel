package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/act"
	bCodes "github.com/sotah-inc/steamwheedle-cartel/pkg/bus/codes"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/metric"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (sta GatewayState) PublishComputedLiveAuctions(tuples sotah.RegionRealmTuples) error {
	// publishing to receive-computed-live-auctions bus endpoint
	jsonEncoded, err := json.Marshal(tuples)
	if err != nil {
		return err
	}

	logging.Info("Publishing to receive-computed-live-auctions bus endpoint")
	req, err := sta.IO.BusClient.Request(sta.receiveComputedLiveAuctionsTopic, string(jsonEncoded), 10*time.Second)
	if err != nil {
		return err
	}

	if req.Code != bCodes.Ok {
		return errors.New(req.Err)
	}

	logging.Info("Finished pushing to receive-computed-live-auctions bus endpoint")

	return nil
}

func (sta GatewayState) PublishToCallSyncAllItems(ids blizzard.ItemIds) error {
	// publishing to call-sync-all-items bus endpoint
	jsonEncoded, err := ids.EncodeForDelivery()
	if err != nil {
		return err
	}

	logging.Info("Publishing to call-sync-all-items bus endpoint")
	req, err := sta.IO.BusClient.Request(sta.callSyncAllItemsTopic, jsonEncoded, 10*time.Second)
	if err != nil {
		return err
	}

	if req.Code != bCodes.Ok {
		return errors.New(req.Err)
	}

	logging.Info("Finished pushing to call-sync-all-items bus endpoint")

	return nil
}

func (sta GatewayState) ComputeLiveAuctionsFromTuples(
	tuples sotah.RegionRealmTimestampTuples,
) (sotah.RegionRealmSummaryTuples, error) {
	// generating new act client
	logging.WithField(
		"endpoint-url",
		sta.actEndpoints.ComputeLiveAuctions,
	).Info("Producing act client compute-live-auctions act endpoint")
	actClient, err := act.NewClient(sta.actEndpoints.ComputeLiveAuctions)
	if err != nil {
		return sotah.RegionRealmSummaryTuples{}, err
	}

	// calling act client with region-realms
	logging.Info("Calling compute-live-auctions with act client")
	nextTuples := sotah.RegionRealmSummaryTuples{}
	actStartTime := time.Now()
	for outJob := range actClient.ComputeLiveAuctions(tuples) {
		// validating that no error occurred during act service calls
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("Failed to compute live-auctions")

			continue
		}

		// handling the job
		switch outJob.Data.Code {
		case http.StatusCreated:
			// parsing the response body
			tuple, err := sotah.NewRegionRealmSummaryTuple(string(outJob.Data.Body))
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
	durationInUs := int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000)
	logging.WithFields(logrus.Fields{
		"duration-in-ms": durationInUs * 1000,
	},
	).Info("Finished calling act compute-live-auctions")

	// reporting metrics
	m := metric.Metrics{
		"compute_all_live_auctions_duration":     int(int64(time.Since(actStartTime)) / 1000 / 1000 / 1000),
		"included_realms_computed_live_auctions": len(tuples),
	}
	if err := sta.IO.BusClient.PublishMetrics(m); err != nil {
		return sotah.RegionRealmSummaryTuples{}, err
	}

	return nextTuples, nil
}

func (sta GatewayState) ComputeAllLiveAuctions(tuples sotah.RegionRealmTimestampTuples) error {
	// computing live-auctions from all region-realms
	nextTuples, err := sta.ComputeLiveAuctionsFromTuples(tuples)
	if err != nil {
		logging.WithField("error", err.Error()).Error("Failed to compute live-auctions from all region-realms")

		return err
	}

	// optionally halting on no results
	if len(tuples) == 0 {
		logging.Info("No realms were updated")

		return nil
	}

	// publishing to receive-realms
	logging.Info("Publishing region-realm tuples to receiver")
	if err := sta.PublishComputedLiveAuctions(nextTuples.RegionRealmTuples()); err != nil {
		return err
	}

	// publishing to sync-items
	logging.Info("Publishing item-ids to call-sync-all-items")
	if err := sta.PublishToCallSyncAllItems(nextTuples.ItemIds()); err != nil {
		return err
	}

	logging.Info("Finished")

	return nil
}
