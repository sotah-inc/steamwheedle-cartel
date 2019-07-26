package act

import (
	"errors"
	"net/http"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func (c Client) CleanupAllAuctions() error {
	actData, err := c.Call("/cleanup-all-auctions", "POST", nil)
	if err != nil {
		return err
	}

	if actData.Code != http.StatusOK {
		logging.WithField("code", actData.Code).Error("Response code was not 200 OK")

		return errors.New("response code was not 200 OK")
	}

	return nil
}
