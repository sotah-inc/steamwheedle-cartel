package database

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

func newLiveAuctionsDatabase(dirPath string, rea sotah.Realm) (liveAuctionsDatabase, error) {
	dbFilepath := liveAuctionsDatabasePath(dirPath, rea)
	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return liveAuctionsDatabase{}, err
	}

	return liveAuctionsDatabase{db, rea}, nil
}

type liveAuctionsDatabase struct {
	db    *bolt.DB
	realm sotah.Realm
}

func (ladBase liveAuctionsDatabase) persistMiniAuctionList(maList sotah.MiniAuctionList) error {
	logging.WithFields(logrus.Fields{
		"db":                 ladBase.db.Path(),
		"mini-auctions-list": len(maList),
	}).Debug("Persisting mini-auction-list")

	encodedData, err := maList.EncodeForDatabase()
	if err != nil {
		return err
	}

	err = ladBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(liveAuctionsBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(liveAuctionsMainKeyName(), encodedData); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ladBase liveAuctionsDatabase) persistEncodedData(encodedData []byte) error {
	logging.WithFields(logrus.Fields{
		"db":           ladBase.db.Path(),
		"encoded-data": len(encodedData),
	}).Debug("Persisting mini-auction-list via encoded-data")

	err := ladBase.db.Update(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(liveAuctionsBucketName())
		if err != nil {
			return err
		}

		if err := bkt.Put(liveAuctionsMainKeyName(), encodedData); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ladBase liveAuctionsDatabase) GetMiniAuctionList() (sotah.MiniAuctionList, error) {
	out := sotah.MiniAuctionList{}

	err := ladBase.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(liveAuctionsBucketName())
		if bkt == nil {
			logging.WithFields(logrus.Fields{
				"db":          ladBase.db.Path(),
				"bucket-name": string(liveAuctionsBucketName()),
			}).Error("Live-auctions bucket not found")

			return nil
		}

		var err error
		out, err = sotah.NewMiniAuctionListFromGzipped(bkt.Get(liveAuctionsMainKeyName()))
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

func NewMiniAuctionListStats(gzipEncoded []byte) (MiniAuctionListStats, error) {
	gzipDecoded, err := util.GzipDecode(gzipEncoded)
	if err != nil {
		return MiniAuctionListStats{}, err
	}

	var jsonDecoded MiniAuctionListStats
	if err := json.Unmarshal(gzipDecoded, &jsonDecoded); err != nil {
		return MiniAuctionListStats{}, err
	}

	return jsonDecoded, nil
}

type MiniAuctionListStats struct {
	TotalAuctions int
	TotalQuantity int
	TotalBuyout   int
	ItemIds       []blizzard.ItemID
	AuctionIds    []int64
}

func (s MiniAuctionListStats) EncodeForStorage() ([]byte, error) {
	jsonEncoded, err := json.Marshal(s)
	if err != nil {
		return []byte{}, err
	}

	return util.GzipEncode(jsonEncoded)
}

func (ladBase liveAuctionsDatabase) stats() (MiniAuctionListStats, error) {
	maList, err := ladBase.GetMiniAuctionList()
	if err != nil {
		return MiniAuctionListStats{}, err
	}

	out := MiniAuctionListStats{
		TotalAuctions: maList.TotalAuctions(),
		TotalQuantity: maList.TotalQuantity(),
		TotalBuyout:   int(maList.TotalBuyout()),
		ItemIds:       maList.ItemIds(),
		AuctionIds:    maList.AuctionIds(),
	}

	return out, nil
}
