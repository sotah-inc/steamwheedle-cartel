package blizzard

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/util"
)

// OAuthTokenEndpoint - http endpoint for gathering new oauth access tokens
const OAuthTokenEndpoint = "https://us.battle.net/oauth/token?grant_type=client_credentials"

// NewClient - generates a client used for querying blizz api
func NewClient(id string, secret string) (Client, error) {
	if len(id) == 0 {
		return Client{}, errors.New("client id is blank")
	}
	if len(secret) == 0 {
		return Client{}, errors.New("client secret is blank")
	}

	initialClient := Client{id, secret, ""}
	client, err := initialClient.RefreshFromHTTP(OAuthTokenEndpoint)
	if err != nil {
		return Client{}, err
	}

	return client, nil
}

// Client - used for querying blizz api
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

// RefreshFromHTTP - gathers an access token from the oauth token endpoint
func (c Client) RefreshFromHTTP(uri string) (Client, error) {
	// forming a request
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return Client{}, err
	}

	// appending auth headers
	req.SetBasicAuth(c.id, c.secret)
	req.Header.Add("Accept-Encoding", "gzip")

	// producing an http client and running the request
	tp := newTimedTransport()
	httpClient := &http.Client{Transport: tp}
	resp, err := httpClient.Do(req)
	if err != nil {
		return Client{}, err
	}

	if resp.StatusCode != http.StatusOK {
		logging.WithFields(logrus.Fields{
			"uri":    uri,
			"id":     c.id,
			"secret": c.secret,
		}).Info("Received failed oauth token response from Blizzard API")

		return Client{}, errors.New("OAuth token response was not 200")
	}

	// parsing the body
	body, isGzipped, err := func() ([]byte, bool, error) {
		defer resp.Body.Close()

		isGzipped := resp.Header.Get("Content-Encoding") == "gzip"
		out, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, false, err
		}

		return out, isGzipped, nil
	}()
	if err != nil {
		return Client{}, err
	}

	// optionally decoding the response body
	decodedBody, err := func() ([]byte, error) {
		if !isGzipped {
			return body, nil
		}

		return util.GzipDecode(body)
	}()
	if err != nil {
		return Client{}, err
	}

	// unmarshalling the body
	r := &refreshResponse{}
	if err := json.Unmarshal(decodedBody, &r); err != nil {
		return Client{}, err
	}

	c.accessToken = r.AccessToken

	return c, nil
}

// AppendAccessToken - appends access token used for making authenticated requests
func (c Client) AppendAccessToken(destination string) (string, error) {
	if c.accessToken == "" {
		return "", errors.New("could not append access token, access token is blank")
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
