package database

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
)

// bucketing
func databasePubsubTopicsBucketName() []byte {
	return []byte("pubsub-topics")
}

// keying
func pubsubTopicsKeyName(topicName string) []byte {
	return []byte(topicName)
}

func pubsubTopicsValueFromTimestamp(t time.Time) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, uint64(t.Unix()))

	return key
}

func topicNameFirstSeen(v []byte) (sotah.UnixTimestamp, error) {
	return sotah.UnixTimestamp(binary.LittleEndian.Uint64(v)), nil
}

// db
func pubsubTopicsDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/pubsub-topics.db", dbDir), nil
}

func topicNamePrefix(projectID string) string {
	return fmt.Sprintf("projects/%s/topics/", projectID)
}
