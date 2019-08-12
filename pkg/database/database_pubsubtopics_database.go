package database

import (
	"github.com/boltdb/bolt"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
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
