package act

import (
	"errors"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/logging"
)

func (c Client) CleanupAllPricelistHistories() error {
	actData, err := c.Call("/cleanup-all-pricelist-histories", "POST", nil)
	if err != nil {
		return err
	}

	if actData.Code != http.StatusOK {
		logging.WithField("code", actData.Code).Error("Response code was not 200 OK")

		return errors.New("response code was not 200 OK")
	}

	return nil
}
