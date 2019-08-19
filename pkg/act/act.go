package act

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/logging"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah"
	"github.com/sotah-inc/steamwheedle-cartel/pkg/sotah/codes"
)

func GetToken(serviceURL string) (string, error) {
	tokenURL := fmt.Sprintf(
		"instance/service-accounts/default/identity?audience=%s",
		serviceURL,
	)
	idToken, err := metadata.Get(tokenURL)
	if err != nil {
		return "", fmt.Errorf("metadata.Get: failed to query id_token: %+v", err)
	}

	return idToken, nil
}

type RequestMeta struct {
	Method     string
	ServiceURL string
	Body       []byte
	Token      string
}

type ResponseMeta struct {
	Body []byte
	Code int
}

func Call(in RequestMeta) (ResponseMeta, error) {
	req, err := http.NewRequest(in.Method, in.ServiceURL, bytes.NewReader(in.Body))
	if err != nil {
		return ResponseMeta{}, err
	}
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", in.Token))

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

func WriteErroneousMessageResponse(w http.ResponseWriter, responseBody string, msg sotah.Message) {
	WriteErroneousResponse(w, codes.CodeToHTTPStatus(msg.Code), responseBody)
}

func WriteErroneousErrorResponse(w http.ResponseWriter, responseBody string, err error) {
	WriteErroneousResponse(w, http.StatusInternalServerError, responseBody)
}

func WriteErroneousResponse(w http.ResponseWriter, code int, responseBody string) {
	if _, err := w.Write([]byte(responseBody)); err != nil {
		logging.WithField("error", err.Error()).Error("Failed to write response")

		return
	}

	w.WriteHeader(code)
}
