package dev_test

import (
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

	stopChan := make(state.ListenStopChan)
	if !assert.Nil(t, apiState.ListenForRealmModificationDates(stopChan)) {
		return
	}

	msg, err := apiState.IO.Messenger.Request(string(subjects.RealmModificationDates), nil)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.Equal(t, codes.Ok, msg.Code) {
		return
	}

	stopChan <- struct{}{}
}