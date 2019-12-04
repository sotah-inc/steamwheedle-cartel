package dev_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/messenger/codes"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state"
	devState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/dev"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/subjects"
)

func TestListenForRealmModificationDates(t *testing.T) {
	assert.Equal(t, 1, 1, "Testing")

	apiState, err := devState.NewAPIState(devState.APIStateConfig{
		SotahConfig: sotah.Config{
			Regions: sotah.RegionList{
				{
					Name:     "us",
					Hostname: "us.api.blizzard.com",
					Primary:  true,
				},
			},
			Whitelist:     map[blizzard.RegionName][]blizzard.RealmSlug{"us": {"earthen-ring"}},
			UseGCloud:     false,
			Expansions:    nil,
			Professions:   nil,
			ItemBlacklist: nil,
		},
		GCloudProjectID:      "",
		MessengerHost:        "localhost",
		MessengerPort:        4222,
		DiskStoreCacheDir:    "/tmp/api-test",
		BlizzardClientId:     os.Getenv("CLIENT_ID"),
		BlizzardClientSecret: os.Getenv("CLIENT_SECRET"),
		ItemsDatabaseDir:     "/tmp/api-test/items",
	})
	if !assert.Nil(t, err) {
		return
	}

	// misc
	targetDates := sotah.RealmModificationDates{Downloaded: 1, LiveAuctionsReceived: 1, PricelistHistoriesReceived: 1}

	// checking that set works
	apiState.RegionRealmModificationDates = apiState.RegionRealmModificationDates.Set(
		"us",
		"earthen-ring",
		targetDates,
	)
	if !assert.Equal(t, targetDates, apiState.RegionRealmModificationDates.Get("us", "earthen-ring")) {
		return
	}

	// starting up the listener
	stopChan := make(state.ListenStopChan)
	if !assert.Nil(t, apiState.ListenForRealmModificationDates(stopChan)) {
		return
	}

	// checking that request works against the expected subject
	msg, err := apiState.IO.Messenger.Request(string(subjects.RealmModificationDates), nil)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.Equal(t, codes.Ok, msg.Code) {
		return
	}

	var result sotah.RegionRealmModificationDates
	if err := json.Unmarshal([]byte(msg.Data), &result); !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, targetDates, result.Get("us", "earthen-ring")) {
		return
	}

	// modifying state and checking the result
	nextTargetDates := sotah.RealmModificationDates{Downloaded: 2, LiveAuctionsReceived: 2, PricelistHistoriesReceived: 2}
	apiState.RegionRealmModificationDates = apiState.RegionRealmModificationDates.Set(
		"us",
		"earthen-ring",
		nextTargetDates,
	)

	// checking that request works against the expected subject
	msg, err = apiState.IO.Messenger.Request(string(subjects.RealmModificationDates), nil)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.Equal(t, codes.Ok, msg.Code) {
		return
	}

	var nextResult sotah.RegionRealmModificationDates
	if err := json.Unmarshal([]byte(msg.Data), &nextResult); !assert.Nil(t, err) {
		return
	}

	if !assert.Equal(t, nextTargetDates, nextResult.Get("us", "earthen-ring")) {
		return
	}

	// cleaning up
	stopChan <- struct{}{}
}
