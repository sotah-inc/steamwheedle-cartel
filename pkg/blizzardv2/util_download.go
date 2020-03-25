package blizzardv2

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/util"
)

type timedTransport struct {
	rtp       http.RoundTripper
	dialer    *net.Dialer
	connStart time.Time
	connEnd   time.Time
	reqStart  time.Time
	reqEnd    time.Time
}

func newTimedTransport() *timedTransport {
	tr := &timedTransport{
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}
	tr.rtp = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         tr.dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return tr
}

func (tr *timedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr.reqStart = time.Now()
	resp, err := tr.rtp.RoundTrip(r)
	tr.reqEnd = time.Now()
	return resp, err
}

func (tr *timedTransport) dial(ctx context.Context, network, addr string) (net.Conn, error) {
	tr.connStart = time.Now()
	cn, err := tr.dialer.DialContext(ctx, network, addr)
	tr.connEnd = time.Now()
	return cn, err
}

func (tr *timedTransport) ReqDuration() time.Duration {
	return tr.Duration() - tr.ConnDuration()
}

func (tr *timedTransport) ConnDuration() time.Duration {
	return tr.connEnd.Sub(tr.connStart)
}

func (tr *timedTransport) Duration() time.Duration {
	return tr.reqEnd.Sub(tr.reqStart)
}

type ResponseError struct {
	Body   string `json:"body"`
	Status int    `json:"status"`
	URI    string `json:"uri"`
}

type ResponseMeta struct {
	ContentLength      int
	Body               []byte
	Status             int
	ConnectionDuration time.Duration
	RequestDuration    time.Duration
	LastModified       int64
}

type DownloadOptions struct {
	Uri             string
	IfModifiedSince time.Time
}

func Download(opts DownloadOptions) (ResponseMeta, error) {
	// forming a request
	req, err := http.NewRequest("GET", opts.Uri, nil)
	if err != nil {
		return ResponseMeta{}, err
	}
	req.Header.Add("Accept-Encoding", "gzip")

	// optionally adding if-modified-since header
	if !opts.IfModifiedSince.IsZero() {
		req.Header.Add("If-Modified-Since", opts.IfModifiedSince.Format(time.RFC1123))
	}

	// running it into a client
	tp := newTimedTransport()
	httpClient := &http.Client{Transport: tp}
	resp, err := httpClient.Do(req)
	if err != nil {
		return ResponseMeta{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logging.WithField("error", err.Error()).Error("Failed to close response body")
		}
	}()

	// checking last-modified header
	parsedLastModified, err := func() (int64, error) {
		foundLastModified := resp.Header.Get("Last-Modified")
		if foundLastModified == "" {
			return 0, nil
		}

		parsed, err := time.Parse(time.RFC1123, foundLastModified)
		if err != nil {
			return 0, err
		}

		return parsed.Unix(), nil
	}()
	if err != nil {
		return ResponseMeta{}, err
	}

	respMeta := ResponseMeta{
		ContentLength:      0,
		Body:               []byte{},
		Status:             resp.StatusCode,
		ConnectionDuration: tp.ConnDuration(),
		RequestDuration:    tp.ReqDuration(),
		LastModified:       parsedLastModified,
	}

	// parsing the body
	body, isGzipped, err := func() ([]byte, bool, error) {
		isGzipped := resp.Header.Get("Content-Encoding") == "gzip"
		out, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, false, err
		}

		return out, isGzipped, nil
	}()
	if err != nil {
		return respMeta, err
	}
	respMeta.ContentLength = len(body)

	// optionally decoding the response body
	decodedBody, err := func() ([]byte, error) {
		if !isGzipped {
			return body, nil
		}

		return util.GzipDecode(body)
	}()
	if err != nil {
		return respMeta, err
	}
	respMeta.Body = decodedBody

	return respMeta, nil
}
