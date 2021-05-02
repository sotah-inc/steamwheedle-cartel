package items

import (
	"github.com/boltdb/bolt"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
)

func (idBase Database) VendorPrice(id blizzardv2.ItemId) (blizzardv2.PriceValue, bool, error) {
	out := blizzardv2.PriceValue(0)
	exists := false

	err := idBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(itemVendorPricesBucket())
		if bkt == nil {
			return nil
		}

		v := bkt.Get(itemVendorPriceKeyName(id))
		if v == nil {
			return nil
		}

		out = itemVendorPriceFromValue(v)
		exists = true

		return nil
	})
	if err != nil {
		return blizzardv2.PriceValue(0), false, err
	}

	return out, exists, nil
}
