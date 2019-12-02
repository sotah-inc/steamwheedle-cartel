package blizzard

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel/pkg/utiltest"
)

func validateItem(i Item) bool {
	return i.ID != 0
}

func TestNewItemFromHTTP(t *testing.T) {
	ts, err := utiltest.ServeFile("../TestData/item.json")
	if !assert.Nil(t, err) {
		return
	}

	a, _, err := NewItemFromHTTP(ts.URL)
	if !assert.Nil(t, err) {
		return
	}
	if !assert.True(t, validateItem(a)) {
		return
	}
}
func TestNewItemFromFilepath(t *testing.T) {
	i, err := NewItemFromFilepath("../TestData/item.json")
	if !assert.Nil(t, err) {
		return
	}
	if !assert.True(t, validateItem(i)) {
		return
	}
}
