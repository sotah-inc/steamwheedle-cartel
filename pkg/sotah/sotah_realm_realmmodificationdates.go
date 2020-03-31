package sotah

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

type RegionRealmModificationDates map[blizzardv2.RegionName]map[blizzardv2.RealmSlug]RealmModificationDates

func (d RegionRealmModificationDates) Get(
	regionName blizzardv2.RegionName,
	realmSlug blizzardv2.RealmSlug,
) RealmModificationDates {
	realmsModDates, ok := d[regionName]
	if !ok {
		return RealmModificationDates{}
	}

	realmModDates, ok := realmsModDates[realmSlug]
	if !ok {
		return RealmModificationDates{}
	}

	return realmModDates
}

func (d RegionRealmModificationDates) Set(
	regionName blizzardv2.RegionName,
	realmSlug blizzardv2.RealmSlug,
	modDates RealmModificationDates,
) RegionRealmModificationDates {
	realmsModDates, ok := d[regionName]
	if !ok {
		logging.WithFields(logrus.Fields{
			"region": regionName,
			"realm":  realmSlug,
		}).Info("Adding new entry to region-realm-mod-dates")

		d[regionName] = map[blizzardv2.RealmSlug]RealmModificationDates{realmSlug: modDates}

		return d
	}

	logging.WithFields(logrus.Fields{
		"region": regionName,
		"realm":  realmSlug,
	}).Info("Setting realm value in existing region in region-realm-mod-dates")

	realmsModDates[realmSlug] = modDates
	d[regionName] = realmsModDates

	return d
}

func (d RegionRealmModificationDates) EncodeForDelivery() ([]byte, error) {
	return json.Marshal(d)
}

type RealmModificationDates struct {
	Downloaded                 int64 `json:"downloaded"`
	LiveAuctionsReceived       int64 `json:"live_auctions_received"`
	PricelistHistoriesReceived int64 `json:"pricelist_histories_received"`
}
