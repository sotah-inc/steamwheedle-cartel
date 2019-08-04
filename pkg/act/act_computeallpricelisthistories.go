package act

import (
	"errors"
	"net/http"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func (c Client) ComputeAllPricelistHistories(tuples sotah.RegionRealmTimestampTuples) error {
	body, err := tuples.EncodeForDelivery()
	if err != nil {
		return err
	}

	actData, err := c.Call("/compute-all-pricelist-histories", "POST", []byte(body))
	if err != nil {
		return err
	}

	if actData.Code != http.StatusCreated {
		logging.WithField("code", actData.Code).Error("Response code was not 201 CREATED")

		return errors.New("response code was not 201 CREATED")
	}

	return nil
}
