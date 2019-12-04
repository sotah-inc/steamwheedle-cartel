package dev_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/blizzard"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/sotah"
	devState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/dev"
)

func TestListenForRealmModificationDates(t *testing.T) {
	assert.Equal(t, 1, 1, "Testing")

	_, err := devState.NewAPIState(devState.APIStateConfig{
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
}
