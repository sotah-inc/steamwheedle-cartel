package items

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2/itemclass"
)

func (idBase Database) GetItemClassItemsMap(
	ids []itemclass.Id,
) (blizzardv2.ItemClassItemsMap, error) {
	out := blizzardv2.NewItemClassItemsMap(ids)

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(itemClassItemsBucket())
		if bkt == nil {
			return nil
		}

		for _, id := range ids {
			value := bkt.Get(itemClassItemsKeyName(id))
			if value == nil {
				continue
			}

			itemIds, err := blizzardv2.NewItemIds(value)
			if err != nil {
				return err
			}

			out[id] = itemIds
		}

		return nil
	})
	if err != nil {
		return blizzardv2.ItemClassItemsMap{}, err
	}

	return out, nil
}

func (idBase Database) PersistItemClassItemsMap(
	iciMap blizzardv2.ItemClassItemsMap,
) error {
	return idBase.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(itemClassItemsBucket())
		if err != nil {
			return err
		}

		for id, itemIds := range iciMap {
			encodedItemIds, err := itemIds.EncodeForDelivery()
			if err != nil {
				return err
			}

			if err := bkt.Put(itemClassItemsKeyName(id), encodedItemIds); err != nil {
				return err
			}
		}

		return nil
	})
}

func (idBase Database) ReceiveItemClassItemsMap(
	iciMap blizzardv2.ItemClassItemsMap,
) error {
	foundIciMap, err := idBase.GetItemClassItemsMap(iciMap.ItemClassIds())
	if err != nil {
		return err
	}

	for id, itemIds := range foundIciMap {
		foundIciMap[id] = itemIds.Merge(iciMap.Find(id))
	}

	return idBase.PersistItemClassItemsMap(foundIciMap)
}

func (idBase Database) GetItemClassItemIds(id itemclass.Id) (blizzardv2.ItemIds, error) {
	out := blizzardv2.ItemIds{}

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(itemClassItemsBucket())
		if bkt == nil {
			return nil
		}

		v := bkt.Get(itemClassItemsKeyName(id))
		if v == nil {
			return nil
		}

		return json.Unmarshal(v, &out)
	})
	if err != nil {
		return blizzardv2.ItemIds{}, err
	}

	return out, nil
}
