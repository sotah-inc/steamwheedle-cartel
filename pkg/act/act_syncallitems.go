package act

import (
	"errors"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

func (c Client) SyncAllItems(ids blizzard.ItemIds) error {
	body, err := ids.EncodeForDelivery()
	if err != nil {
		return err
	}

	actData, err := c.Call("/sync-all-items", "POST", []byte(body))
	if err != nil {
		return err
	}

	if actData.Code != http.StatusCreated {
		logging.WithField("code", actData.Code).Error("Response code was not 201 CREATED")

		return errors.New("response code was not 201 CREATED")
	}

	return nil
}
