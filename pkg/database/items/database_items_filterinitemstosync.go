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

func (idBase Database) FilterInItemsToSync(ids []blizzardv2.ItemId) (SyncPayload, error) {
	// producing a blank whitelist
	syncWhitelist := sotah.NewItemSyncWhitelist(ids)

	// producing a blank map of icon->item-ids
	iconsToSync := sotah.IconIdsMap{}

	// peeking into the items database
	err := idBase.db.View(func(tx *bolt.Tx) error {
		itemsBucket := tx.Bucket(baseBucketName())
		if itemsBucket == nil {
			return nil
		}

		itemNamesBucket := tx.Bucket(namesBucketName())
		if itemNamesBucket == nil {
			return nil
		}

		for _, id := range ids {
			value := itemsBucket.Get(baseKeyName(id))
			if value == nil {
				syncWhitelist[id] = true

				continue
			}

			item, err := sotah.NewItemFromGzipped(value)
			if err != nil {
				return err
			}

			hasBlankIconMeta := item.SotahMeta.ItemIconMeta.IsZero()
			hasIncorrectIconMeta := func() bool {
				if hasBlankIconMeta {
					return false
				}

				correctIconObjectName := sotah.NewItemObjectName(item.SotahMeta.ItemIconMeta.Icon)
				//correctIconURL := fmt.Sprintf(
				//	store.ItemIconURLFormat,
				//	"sotah-item-icons",
				//	correctIconObjectName,
				//)
				correctIconURL := blizzardv2.DefaultGetItemIconURL(correctIconObjectName)

				return item.SotahMeta.ItemIconMeta.ObjectName != correctIconObjectName ||
					item.SotahMeta.ItemIconMeta.URL != correctIconURL
			}()
			if hasBlankIconMeta || hasIncorrectIconMeta {
				iconsToSync = iconsToSync.Append(item.SotahMeta.ItemIconMeta.Icon, item.BlizzardMeta.Id)
			}

			isMissingNames := item.SotahMeta.NormalizedName.IsZero()
			isMissingNormalizedName := itemNamesBucket.Get(nameKeyName(id)) == nil
			if isMissingNames || isMissingNormalizedName {
				syncWhitelist[id] = true
			}
		}

		return nil
	})
	if err != nil {
		return SyncPayload{}, err
	}

	return SyncPayload{Ids: syncWhitelist.ToItemIds(), IconIdsMap: iconsToSync}, nil
}
