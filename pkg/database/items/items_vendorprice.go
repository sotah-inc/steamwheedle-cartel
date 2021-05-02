package items

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzardv2"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
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

type VendorPricesJob struct {
	Err         error
	Id          blizzardv2.ItemId
	Exists      bool
	VendorPrice blizzardv2.PriceValue
}

func (job VendorPricesJob) ToLogrusFields() logrus.Fields {
	return logrus.Fields{
		"error": job.Err.Error(),
		"item":  job.Id,
	}
}

func (idBase Database) VendorPrices(
	ids blizzardv2.ItemIds,
) (map[blizzardv2.ItemId]blizzardv2.PriceValue, error) {
	in := make(chan blizzardv2.ItemId)
	out := make(chan VendorPricesJob)
	worker := func() {
		for id := range in {
			vendorPrice, exists, err := idBase.VendorPrice(id)
			if err != nil {
				out <- VendorPricesJob{
					Err:         err,
					Id:          id,
					Exists:      false,
					VendorPrice: 0,
				}

				continue
			}

			out <- VendorPricesJob{
				Err:         nil,
				Id:          id,
				Exists:      exists,
				VendorPrice: vendorPrice,
			}
		}
	}
	postWork := func() {
		close(out)
	}
	util.Work(4, worker, postWork)

	// queueing it up
	go func() {
		for _, id := range ids {
			in <- id
		}

		close(in)
	}()

	results := map[blizzardv2.ItemId]blizzardv2.PriceValue{}
	for outJob := range out {
		if outJob.Err != nil {
			logging.WithFields(outJob.ToLogrusFields()).Error("failed to resolve item vendor-price")

			return map[blizzardv2.ItemId]blizzardv2.PriceValue{}, outJob.Err
		}

		if !outJob.Exists {
			continue
		}

		results[outJob.Id] = outJob.VendorPrice
	}

	return results, nil
}
