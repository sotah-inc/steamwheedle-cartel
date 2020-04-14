package database

import (
	"encoding/base64"
	"encoding/json"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func NewItemsSyncPayload(data string) (ItemsSyncPayload, error) {
	base64Decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return ItemsSyncPayload{}, err
	}

	gzipDecoded, err := util.GzipDecode(base64Decoded)
	if err != nil {
		return ItemsSyncPayload{}, err
	}

	var out ItemsSyncPayload
	if err := json.Unmarshal(gzipDecoded, &out); err != nil {
		return ItemsSyncPayload{}, err
	}

	return out, nil
}

type ItemsSyncPayload struct {
	Ids        blizzardv2.ItemIds
	IconIdsMap sotah.IconIdsMap
}

func (p ItemsSyncPayload) EncodeForDelivery() (string, error) {
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

func (idBase ItemsDatabase) FilterInItemsToSync(ids []blizzardv2.ItemId) (ItemsSyncPayload, error) {
	// producing a blank whitelist
	syncWhitelist := sotah.NewItemSyncWhitelist(ids)

	// producing a blank map of icon->item-ids
	iconsToSync := sotah.IconIdsMap{}

	// peeking into the items database
	err := idBase.db.Update(func(tx *bolt.Tx) error {
		itemsBucket, err := tx.CreateBucketIfNotExists(databaseItemsBucketName())
		if err != nil {
			return err
		}

		itemNamesBucket, err := tx.CreateBucketIfNotExists(databaseItemNamesBucketName())
		if err != nil {
			return err
		}

		for _, id := range ids {
			value := itemsBucket.Get(itemsKeyName(id))
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
				//correctIconURL := fmt.Sprintf(store.ItemIconURLFormat, "sotah-item-icons", correctIconObjectName)
				correctIconURL := blizzardv2.DefaultGetItemIconURL(correctIconObjectName)

				return item.SotahMeta.ItemIconMeta.ObjectName != correctIconObjectName ||
					item.SotahMeta.ItemIconMeta.URL != correctIconURL
			}()
			if hasBlankIconMeta || hasIncorrectIconMeta {
				iconsToSync = iconsToSync.Append(item.SotahMeta.ItemIconMeta.Icon, item.BlizzardMeta.Id)
			}

			isMissingNames := item.SotahMeta.NormalizedName.IsZero()
			isMissingNormalizedName := itemNamesBucket.Get(itemNameKeyName(id)) == nil
			if isMissingNames || isMissingNormalizedName {
				syncWhitelist[id] = true
			}
		}

		return nil
	})
	if err != nil {
		return ItemsSyncPayload{}, err
	}

	return ItemsSyncPayload{Ids: syncWhitelist.ToItemIds(), IconIdsMap: iconsToSync}, nil
}
