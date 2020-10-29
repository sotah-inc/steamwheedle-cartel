package liveauctions

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
)

func (ladBase LiveAuctionsDatabase) persistEncodedData(encodedData []byte) error {
	logging.WithFields(logrus.Fields{
		"db":           ladBase.db.Path(),
		"encoded-data": len(encodedData),
	}).Debug("persisting mini-auction-list via encoded-data")

	err := ladBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(baseBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(baseKeyName(), encodedData); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ladBase LiveAuctionsDatabase) GetMiniAuctionList() (sotah.MiniAuctionList, error) {
	out := sotah.MiniAuctionList{}

	err := ladBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(baseBucketName())
		if bkt == nil {
			logging.WithFields(logrus.Fields{
				"db":          ladBase.db.Path(),
				"bucket-name": string(baseBucketName()),
			}).Error("live-auctions bucket not found")

			return nil
		}

		var err error
		out, err = sotah.NewMiniAuctionListFromGzipped(bkt.Get(baseKeyName()))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return sotah.MiniAuctionList{}, err
	}

	return out, nil
}
