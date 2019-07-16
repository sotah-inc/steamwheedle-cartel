package act

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
)

func Create(url string, body io.Reader) (ResponseMeta, error) {
	return Call("POST", url, body)
}

func Update(url string, body io.Reader) (ResponseMeta, error) {
	return Call("PUT", url, body)
}

func Get(url string, body io.Reader) (ResponseMeta, error) {
	return Call("Get", url, body)
}

type ResponseMeta struct {
	Body []byte
	Code int
}

func Call(method string, url string, body io.Reader) (ResponseMeta, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return ResponseMeta{}, err
	}
	req.Header.Add("Accept-Encoding", "gzip")

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
