package database

import (
	"time"

	"github.com/boltdb/bolt"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

func NewPubsubTopicsDatabase(dbDir string) (PubsubTopicsDatabase, error) {
	dbFilepath, err := pubsubTopicsDatabasePath(dbDir)
	if err != nil {
		return PubsubTopicsDatabase{}, err
	}

	logging.WithField("filepath", dbFilepath).Info("Initializing pubsub-topics database")

	db, err := bolt.Open(dbFilepath, 0600, nil)
	if err != nil {
		return PubsubTopicsDatabase{}, err
	}

	return PubsubTopicsDatabase{db}, nil
}

type PubsubTopicsDatabase struct {
	db *bolt.DB
}

func NewTopicNamesFirstSeen(topicNames []string) TopicNamesFirstSeen {
	out := map[string]sotah.UnixTimestamp{}
	for _, name := range topicNames {
		out[name] = sotah.UnixTimestamp(0)
	}

	return out
}

type TopicNamesFirstSeen map[string]sotah.UnixTimestamp

func (b PubsubTopicsDatabase) Current(topicNames []string) (TopicNamesFirstSeen, error) {
	out := NewTopicNamesFirstSeen(topicNames)

	err := b.db.View(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(databasePubsubTopicsBucketName())
		if err != nil {
			return err
		}

		err = bkt.ForEach(func(k, v []byte) error {
			firstSeenTimestamp, err := topicNameFirstSeen(v)
			if err != nil {
				return err
			}

			out[string(k)] = firstSeenTimestamp

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return map[string]sotah.UnixTimestamp{}, err
	}

	return out, nil
}

func (b PubsubTopicsDatabase) Fill(topicNames []string, currentTime time.Time) error {
	return nil
}
