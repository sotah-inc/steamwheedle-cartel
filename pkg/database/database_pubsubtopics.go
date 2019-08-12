package database

import (
	"fmt"
)

// bucketing
func databasePubsubTopicsBucketName() []byte {
	return []byte("pubsub-topics")
}

// keying
func pubsubTopicsKeyName(topicName string) []byte {
	return []byte(fmt.Sprintf("pubsub-topics-%s", topicName))
}

// db
func pubsubTopicsDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/pubsub-topics.db", dbDir), nil
}
