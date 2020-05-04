package disk

import (
	"errors"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func (c Client) Collect() (blizzardv2.ItemIds, error) {
	startTime := time.Now()
	logging.Info("calling DiskCollector.Collect()")

	results, err := c.collectAuctions()
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	if err := c.CallLiveAuctionsIntake(results.tuples.RegionConnectedRealmTuples()); err != nil {
		return blizzardv2.ItemIds{}, err
	}

	if err := c.CallPricelistHistoryIntake(results.tuples); err != nil {
		return blizzardv2.ItemIds{}, err
	}

	logging.WithField(
		"duration-in-ms",
		time.Since(startTime).Milliseconds(),
	).Info("finished calling DiskCollector.Collect()")

	return results.itemIds, nil
}

func (c Client) CallLiveAuctionsIntake(tuples blizzardv2.RegionConnectedRealmTuples) error {
	// forwarding the received tuples to live-auctions intake
	encodedTuples, err := tuples.EncodeForDelivery()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to encode load tuples for delivery")

		return err
	}

	response, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.LiveAuctionsIntake),
		Data:    encodedTuples,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for pricelist-history intake",
		)

		return err
	}

	if response.Code != codes.Ok {
		logging.WithFields(response.ToLogrusFields()).Error("pricelist-history intake request failed")

		return errors.New(response.Err)
	}

	return nil
}

func (c Client) CallPricelistHistoryIntake(tuples blizzardv2.LoadConnectedRealmTuples) error {
	// forwarding the received tuples to pricelist-history intake
	encodedTuples, err := tuples.EncodeForDelivery()
	if err != nil {
		logging.WithField("error", err.Error()).Error("failed to encode load tuples for delivery")

		return err
	}

	response, err := c.messengerClient.Request(messenger.RequestOptions{
		Subject: string(subjects.PricelistHistoryIntake),
		Data:    encodedTuples,
		Timeout: 10 * time.Minute,
	})
	if err != nil {
		logging.WithField("error", err.Error()).Error(
			"failed to publish message for pricelist-history intake",
		)

		return err
	}

	if response.Code != codes.Ok {
		logging.WithFields(response.ToLogrusFields()).Error("pricelist-history intake request failed")

		return errors.New(response.Err)
	}

	return nil
}
