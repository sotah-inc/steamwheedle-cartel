package act

import (
	"errors"
	"net/http"
)

func (c Client) DownloadAllAuctions() error {
	actData, err := c.Call("/download-all-auctions", "POST", nil)
	if err != nil {
		return err
	}

	if actData.Code != http.StatusCreated {
		return errors.New("response code was not 201 CREATED")
	}

	return nil
}
