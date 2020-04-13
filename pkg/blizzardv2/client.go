package blizzardv2

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

const OAuthTokenEndpoint = "https://us.battle.net/oauth/token?grant_type=client_credentials"

type ClientConfig struct {
	ClientId     string
	ClientSecret string
}

func NewClient(config ClientConfig) (Client, error) {
	if len(config.ClientId) == 0 {
		return Client{}, errors.New("client id is blank")
	}
	if len(config.ClientSecret) == 0 {
		return Client{}, errors.New("client secret is blank")
	}

	client := Client{
		id:          config.ClientId,
		secret:      config.ClientSecret,
		accessToken: "",
	}
	if err := client.RefreshFromHTTP(OAuthTokenEndpoint); err != nil {
		return Client{}, err
	}

	if !client.IsValid() {
		logging.WithField("source", "NewClient").Error("client was not valid")

		return Client{}, errors.New("client was not valid")
	}

	return client, nil
}

type Client struct {
	id          string
	secret      string
	accessToken string
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (c Client) RefreshFromHTTP(uri string) error {
	// forming a request
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}

	// appending auth headers
	req.SetBasicAuth(c.id, c.secret)
	req.Header.Add("Accept-Encoding", "gzip")

	// producing an http client and running the request
	tp := newTimedTransport()
	httpClient := &http.Client{Transport: tp}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"uri":    uri,
			"id":     c.id,
			"secret": c.secret,
		}).Info("Received failed oauth token response from Blizzard API")

		return errors.New("OAuth token response was not 200")
	}

	// parsing the body
	body, isGzipped, err := func() ([]byte, bool, error) {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to close response body")
			}
		}()

		isGzipped := resp.Header.Get("Content-Encoding") == "gzip"
		out, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, false, err
		}

		return out, isGzipped, nil
	}()
	if err != nil {
		return err
	}

	// optionally decoding the response body
	decodedBody, err := func() ([]byte, error) {
		if !isGzipped {
			return body, nil
		}

		return util.GzipDecode(body)
	}()
	if err != nil {
		return err
	}

	// unmarshalling the body
	r := &refreshResponse{}
	if err := json.Unmarshal(decodedBody, &r); err != nil {
		return err
	}

	c.accessToken = r.AccessToken

	return nil
}

func (c Client) AppendAccessToken(destination string) (string, error) {
	if c.accessToken == "" {
		return "", errors.New("could not append access token, access token is a blank string")
	}

	u, err := url.Parse(destination)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", c.accessToken)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c Client) IsValid() bool {
	return c.accessToken != ""
}

func ClearAccessToken(destination string) (string, error) {
	u, err := url.Parse(destination)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("access_token", "xxx")
	u.RawQuery = q.Encode()

	return u.String(), nil
}
