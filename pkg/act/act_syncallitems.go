package act

import (
	"errors"
	"net/http"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/blizzard"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
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
