package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/gameversion"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

// nolint:lll
const connectedRealmIndexURLFormat = "https://%s/data/wow/connected-realm/index?namespace=%s"

func DefaultConnectedRealmIndexURL(
	regionHostname string,
	version gameversion.GameVersion,
) (string, error) {
	namespace, err := gameversion.DynamicNamespaceMap.Resolve(version)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		connectedRealmIndexURLFormat,
		regionHostname,
		namespace,
	), nil
}

type GetConnectedRealmIndexURLFunc func(string) string

type ConnectedRealmIndexResponse struct {
	LinksBase
	ConnectedRealms []HrefReference `json:"connected_realms"`
}

func NewConnectedRealmIndexFromHTTP(uri string) (ConnectedRealmIndexResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download connected-realm-index")

		return ConnectedRealmIndexResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from connected-realm-index was not 200")

		return ConnectedRealmIndexResponse{}, resp, errors.New("status was not 200")
	}

	crIndex, err := NewConnectedRealmIndexResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse connected-realm-index response")

		return ConnectedRealmIndexResponse{}, resp, err
	}

	return crIndex, resp, nil
}

func NewConnectedRealmIndexResponse(body []byte) (ConnectedRealmIndexResponse, error) {
	crIndex := &ConnectedRealmIndexResponse{}
	if err := json.Unmarshal(body, crIndex); err != nil {
		return ConnectedRealmIndexResponse{}, err
	}

	return *crIndex, nil
}
