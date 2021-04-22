package items

import (
	"encoding/base64"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewSyncPayload(data string) (SyncPayload, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return SyncPayload{}, err
	}

	gzipDecoded, err := util.GzipDecode(base64Decoded)
	if err != nil {
		return SyncPayload{}, err
	}

	var out SyncPayload
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return SyncPayload{}, err
	}

	return out, nil
}

type SyncPayload struct {
	Ids        blizzardv2.ItemIds
	IconIdsMap sotah.IconIdsMap
}

func (p SyncPayload) EncodeForDelivery() (string, error) {
	jsonEncoded, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	gzipEncoded, err := util.GzipEncode(jsonEncoded)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(gzipEncoded), nil
}

func (idBase Database) FilterInItemsToSync(
	providedIds blizzardv2.ItemIds,
) (blizzardv2.ItemIds, error) {
	// gathering blacklisted ids
	blacklistedIds, err := idBase.GetBlacklistedIds()
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	logging.WithFields(logrus.Fields{
		"provided-items":    len(providedIds),
		"blacklisted-items": len(blacklistedIds),
	}).Info("filtering blacklisted from provided")

	nextItemIds := providedIds.Sub(blacklistedIds)
	if len(nextItemIds) == 0 {
		logging.Info("resulting next-items was blank, skipping early")

		return blizzardv2.ItemIds{}, nil
	}

	// producing a blank whitelist
	syncWhitelist := sotah.NewItemSyncWhitelist(nextItemIds)

	// peeking into the items database
	logging.Info("checking items database for items to sync")
	err = idBase.db.View(func(tx *bolt.Tx) error {
		itemsBucket := tx.Bucket(baseBucketName())
		if itemsBucket == nil {
			syncWhitelist = syncWhitelist.ActivateAll()

			return nil
		}

		itemNamesBucket := tx.Bucket(namesBucketName())
		if itemNamesBucket == nil {
			syncWhitelist = syncWhitelist.ActivateAll()

			return nil
		}

		for _, id := range nextItemIds {
			value := itemsBucket.Get(baseKeyName(id))
			if value != nil {
				continue
			}

			syncWhitelist[id] = true
		}

		return nil
	})
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	syncItems := syncWhitelist.ToItemIds()

	logging.WithField(
		"sync-whitelist-items",
		len(syncItems),
	).Info("returning sync-whitelist items")

	return syncItems, nil
}
