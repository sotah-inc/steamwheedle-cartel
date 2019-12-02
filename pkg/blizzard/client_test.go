package blizzard

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/utiltest"
)

func TestClientRefresh(t *testing.T) {
	client, err := NewClient("", "")
	if !assert.Nil(t, err) {
		return
	}

	ts, err := utiltest.ServeFile("../TestData/access-token.json")
	if !assert.Nil(t, err) {
		return
	}

	client, err = client.RefreshFromHTTP(ts.URL)
	if !assert.Nil(t, err) {
		return
	}
	if assert.Equal(t, "xxx", client.accessToken) {
		return
	}
}

func TestAppendAccessToken(t *testing.T) {
	client, err := NewClient("", "")
	if !assert.Nil(t, err) {
		return
	}

	ts, err := utiltest.ServeFile("../TestData/access-token.json")
	if !assert.Nil(t, err) {
		return
	}

	client, err = client.RefreshFromHTTP(ts.URL)
	if !assert.Nil(t, err) {
		return
	}

	dest, err := client.AppendAccessToken("https://google.ca/")
	if !assert.Nil(t, err) {
		return
	}
	if !assert.Equal(t, "https://google.ca/?access_token=xxx", dest) {
		return
	}
}
func TestAppendAccessTokenFail(t *testing.T) {
	client, err := NewClient("", "")
	if !assert.Nil(t, err) {
		return
	}

	if _, err := client.AppendAccessToken("https://google.ca/"); !assert.NotNil(t, err) {
		return
	}
}
