package blizzardv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/locale"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/realmpopulation" // nolint:lll
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/realmstatus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

const connectedRealmURLFormat = "https://%s/data/wow/connected-realm/%d?namespace=dynamic-%s"

func DefaultConnectedRealmURL(
	regionHostname string,
	regionName RegionName,
	id ConnectedRealmId,
) string {
	return fmt.Sprintf(connectedRealmURLFormat, regionHostname, id, regionName)
}

type GetConnectedRealmURLFunc func(string, RegionName, ConnectedRealmId) string

type ConnectedRealmId int

type ConnectedRealmResponse struct {
	LinksBase
	Id       ConnectedRealmId `json:"id"`
	HasQueue bool             `json:"has_queue"`
	Status   struct {
		Type realmstatus.RealmStatus `json:"type"`
		Name locale.Mapping          `json:"name"`
	} `json:"status"`
	Population struct {
		Type realmpopulation.RealmPopulation `json:"type"`
		Name locale.Mapping                  `json:"name"`
	} `json:"population"`
	Realms             RealmResponses `json:"realms"`
	MythicLeaderboards HrefReference  `json:"mythic_leaderboards"`
	Auctions           HrefReference  `json:"auctions"`
}

func NewConnectedRealmResponseFromHTTP(uri string) (ConnectedRealmResponse, ResponseMeta, error) {
	resp, err := Download(DownloadOptions{Uri: uri})
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to download connected-realm")

		return ConnectedRealmResponse{}, resp, err
	}

	if resp.Status != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"status": resp.Status,
			"uri":    ClearAccessToken(uri),
		}).Error("resp from connected-realm was not 200")

		return ConnectedRealmResponse{}, resp, errors.New("status was not 200")
	}

	cRealm, err := NewConnectedRealmResponse(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"error": err.Error(),
			"uri":   ClearAccessToken(uri),
		}).Error("failed to parse connected-realm response")

		return ConnectedRealmResponse{}, resp, err
	}

	return cRealm, resp, nil
}

func NewConnectedRealmResponse(body []byte) (ConnectedRealmResponse, error) {
	cRealm := &ConnectedRealmResponse{}
	if err := json.Unmarshal(body, cRealm); err != nil {
		return ConnectedRealmResponse{}, err
	}

	return *cRealm, nil
}
