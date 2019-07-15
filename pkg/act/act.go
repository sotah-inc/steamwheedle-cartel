package act

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func Create(url string, body io.Reader) ([]byte, error) {
	return Call("POST", url, body)
}

func Update(url string, body io.Reader) ([]byte, error) {
	return Call("PUT", url, body)
}

func Call(method string, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Accept-Encoding", "gzip")

	// running it into a client
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logging.WithField("error", err.Error()).Error("Failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("response was not OK: %d", resp.StatusCode)
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		defer func() {
			if err := reader.Close(); err != nil {
				logging.WithField("error", err.Error()).Error("Failed to close gzip body reader")
			}
		}()

		out, err := ioutil.ReadAll(reader)
		if err != nil {
			return []byte{}, err
		}

		return out, nil
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return out, nil
}
