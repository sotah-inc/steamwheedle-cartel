package items

import (
	"encoding/base64"
	"encoding/json"

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

	nextItemIds := providedIds.Sub(blacklistedIds)
	if len(nextItemIds) == 0 {
		return blizzardv2.ItemIds{}, nil
	}

	// producing a blank whitelist
	syncWhitelist := sotah.NewItemSyncWhitelist(nextItemIds)

	// peeking into the items database
	err = idBase.db.View(func(tx *bolt.Tx) error {
		itemsBucket := tx.Bucket(baseBucketName())
		if itemsBucket == nil {
			return nil
		}

		itemNamesBucket := tx.Bucket(namesBucketName())
		if itemNamesBucket == nil {
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

	return syncWhitelist.ToItemIds(), nil
}
