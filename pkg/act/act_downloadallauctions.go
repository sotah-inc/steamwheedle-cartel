package act

import (
	"errors"
	"net/http"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func (c Client) DownloadAllAuctions() error {
	actData, err := c.Call("/download-all-auctions", "POST", nil)
	if err != nil {
		return err
	}

	if actData.Code != http.StatusCreated {
		logging.WithField("code", actData.Code).Error("Response code was not 201 CREATED")

		return errors.New("response code was not 201 CREATED")
	}

	return nil
}
