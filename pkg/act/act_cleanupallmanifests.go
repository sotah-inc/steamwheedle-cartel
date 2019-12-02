package act

import (
	"errors"
	"net/http"

	"git.sotah.info/steamwheedle-cartel/pkg/logging"
)

func (c Client) CleanupAllManifests() error {
	actData, err := c.Call("/cleanup-all-manifests", "POST", nil)
	if err != nil {
		return err
	}

	if actData.Code != http.StatusOK {
		logging.WithField("code", actData.Code).Error("Response code was not 200 OK")

		return errors.New("response code was not 200 OK")
	}

	return nil
}
