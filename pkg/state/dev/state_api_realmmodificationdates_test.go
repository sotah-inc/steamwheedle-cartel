package dev_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	devState "source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/state/dev"
)

func TestListenForRealmModificationDates(t *testing.T) {
	assert.Equal(t, 1, 1, "Testing")

	_, err := devState.NewAPIState(devState.APIStateConfig{})
	if !assert.Nil(t, err) {
		return
	}
}
