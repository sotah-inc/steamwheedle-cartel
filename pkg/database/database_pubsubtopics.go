package database

import (
	"encoding/binary"
	"fmt"
	"strconv"
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
	firstSeenTimestamp, err := strconv.Atoi(string(v))
	if err != nil {
		return 0, err
	}

	return sotah.UnixTimestamp(firstSeenTimestamp), nil
}

// db
func pubsubTopicsDatabasePath(dbDir string) (string, error) {
	return fmt.Sprintf("%s/pubsub-topics.db", dbDir), nil
}

func topicNamePrefix(projectID string) string {
	return fmt.Sprintf("projects/%s/topics/", projectID)
}
