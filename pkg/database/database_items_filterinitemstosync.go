package database

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah/gameversions"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/store"
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
	Ids        []blizzardv2.ItemId
	IconIdsMap map[string][]blizzardv2.ItemId
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
	syncWhitelist := map[blizzardv2.ItemId]bool{}
	for _, id := range ids {
		syncWhitelist[id] = false
	}

	// producing a blank map of icon->item-ids
	iconsToSync := map[string][]blizzardv2.ItemId{}

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
				logging.WithField("item", id).Info("Item was not in bucket")
				syncWhitelist[id] = true

				continue
			}

			item, err := sotah.NewItemFromGzipped(value)
			if err != nil {
				return err
			}

			if item.SotahMeta.ItemIconMeta.IsZero() {
				iconsToSync[item.SotahMeta.ItemIconMeta.Icon] = item.BlizzardMeta.Id
			}
			if !item.SotahMeta.ItemIconMeta.IsZero() {
				correctIconObjectName := fmt.Sprintf(
					"%s/%s.jpg",
					gameversions.Retail,
					item.SotahMeta.ItemIconMeta.ObjectName,
				)
				correctIconURL := fmt.Sprintf(store.ItemIconURLFormat, "sotah-item-icons", correctIconObjectName)

				shouldInclude := item.SotahMeta.ItemIconMeta.ObjectName != correctIconObjectName ||
					item.SotahMeta.ItemIconMeta.URL != correctIconURL
				if shouldInclude {
					iconItemIds := func() []blizzardv2.ItemId {
						out, ok := iconsToSync[item.Icon]
						if !ok {
							return []blizzardv2.ItemId{}
						}

						return out
					}()
					iconItemIds = append(iconItemIds, id)
					iconsToSync[item.Icon] = iconItemIds
				}
			}

			if item.SotahMeta.NormalizedName.IsZero() {
				logging.WithField("item", item.BlizzardMeta.Id).Info("Normalized-name is blank")
				syncWhitelist[id] = true
			}

			normalizedNameValue := itemNamesBucket.Get(itemNameKeyName(id))
			if normalizedNameValue == nil {
				logging.WithField("item", item.ID).Info("Normalized-name not in bucket")
				syncWhitelist[id] = true
			} else {
				if string(normalizedNameValue) == "" {
					logging.WithField("item", item.ID).Info("Normalized-name was a blank string")
					syncWhitelist[id] = true
				} else {
					if string(normalizedNameValue) != item.NormalizedName {
						logging.WithField("item", item.ID).Info("Normalized-name did not match item normalized-name")
						syncWhitelist[id] = true
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return ItemsSyncPayload{}, err
	}

	// reformatting the whitelist
	idsToSync := []blizzardv2.ItemId{}
	for id, shouldSync := range syncWhitelist {
		if !shouldSync {
			continue
		}

		idsToSync = append(idsToSync, id)
	}

	return ItemsSyncPayload{Ids: idsToSync, IconIdsMap: iconsToSync}, nil
}
