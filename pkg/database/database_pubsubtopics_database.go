package database

import (
	"time"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
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

func (s TopicNamesFirstSeen) NonZero() TopicNamesFirstSeen {
	out := TopicNamesFirstSeen{}
	for k, v := range s {
		if v == 0 {
			continue
		}

		out[k] = v
	}

	return out
}

func (s TopicNamesFirstSeen) After(limit time.Time) TopicNamesFirstSeen {
	out := TopicNamesFirstSeen{}
	for k, v := range s {
		if time.Unix(int64(v), 0).After(limit) {
			continue
		}

		out[k] = v
	}

	return out
}

func (s TopicNamesFirstSeen) Names() []string {
	out := make([]string, len(s))
	i := 0
	for k := range s {
		out[i] = k

		i++
	}

	return out
}

func (b PubsubTopicsDatabase) Current(topicNames []string) (TopicNamesFirstSeen, error) {
	out := NewTopicNamesFirstSeen(topicNames)

	err := b.db.Batch(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(databasePubsubTopicsBucketName()); err != nil {
			return err
		}

		return nil
	})

	err = b.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(databasePubsubTopicsBucketName())

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

func (b PubsubTopicsDatabase) Fill(topicNames []string, currentTime time.Time) (TopicNamesFirstSeen, error) {
	currentSeen, err := b.Current(topicNames)
	if err != nil {
		return TopicNamesFirstSeen{}, err
	}

	logging.WithFields(logrus.Fields{
		"current-seen": len(currentSeen.NonZero()),
		"total-seen":   len(currentSeen),
	}).Info("Topic-names provided")

	for k, v := range currentSeen {
		if v == 0 {
			currentSeen[k] = sotah.UnixTimestamp(currentTime.Unix())
		}
	}

	err = b.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(databasePubsubTopicsBucketName())
		if err != nil {
			return err
		}

		for k, v := range currentSeen {
			if err := bkt.Put(pubsubTopicsKeyName(k), pubsubTopicsValueFromTimestamp(time.Unix(int64(v), 0))); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return TopicNamesFirstSeen{}, err
	}

	return currentSeen, nil
}

func (b PubsubTopicsDatabase) Clean(topicNames []string) error {
	err := b.db.Batch(func(tx *bolt.Tx) error {
		bkt, err := tx.CreateBucketIfNotExists(databasePubsubTopicsBucketName())
		if err != nil {
			return err
		}

		for _, topicName := range topicNames {
			if err := bkt.Delete(pubsubTopicsKeyName(topicName)); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
