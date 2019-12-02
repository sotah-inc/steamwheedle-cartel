package act

import (
	"errors"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/sotah"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/logging"
)

func (c Client) ComputeAllLiveAuctions(tuples sotah.RegionRealmTimestampTuples) error {
	body, err := tuples.EncodeForDelivery()
	if err != nil {
		return err
	}

	actData, err := c.Call("/compute-all-live-auctions", "POST", []byte(body))
	if err != nil {
		return err
	}

	if actData.Code != http.StatusCreated {
		logging.WithField("code", actData.Code).Error("Response code was not 201 CREATED")

		return errors.New("response code was not 201 CREATED")
	}

	return nil
}
