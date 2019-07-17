package act

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"

	"cloud.google.com/go/compute/metadata"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func GetToken(serviceURL string) (string, error) {
	tokenURL := fmt.Sprintf(
		"http://metadata/computeMetadata/v1/instance/service-accounts/default/identityit status?audience=%s",
		serviceURL,
	)

	logging.WithField("token-url", tokenURL).Info("Fetching id-token from metadata api")

	idToken, err := metadata.Get(tokenURL)
	if err != nil {
		return "", fmt.Errorf("metadata.Get: failed to query id_token: %+v", err)
	}

	logging.WithField("token", idToken).Info("Received token")

	return idToken, nil
}

type RequestMeta struct {
	Method     string
	ServiceURL string
	Body       io.Reader
	Token      string
}

type ResponseMeta struct {
	Body []byte
	Code int
}

func Call(in RequestMeta) (ResponseMeta, error) {
	req, err := http.NewRequest(in.Method, in.ServiceURL, in.Body)
	if err != nil {
		return ResponseMeta{}, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", in.Token))

	logging.WithFields(logrus.Fields{
		"authorization": fmt.Sprintf("Bearer %s", in.Token),
		"url":           in.ServiceURL,
	}).Info("Calling with token")

	// running it into a client
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return ResponseMeta{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logging.WithField("error", err.Error()).Error("Failed to close response body")
		}
	}()

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return ResponseMeta{}, err
		}
		defer func() {
			if err := reader.Close(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to close gzip body reader")
			}
		}()

		out, err := ioutil.ReadAll(reader)
		if err != nil {
			return ResponseMeta{}, err
		}

		return ResponseMeta{Body: out, Code: resp.StatusCode}, nil
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResponseMeta{}, err
	}

	return ResponseMeta{Body: out, Code: resp.StatusCode}, nil
}
