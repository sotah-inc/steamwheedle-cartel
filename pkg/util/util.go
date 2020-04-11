package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"source.developers.google.com/p/sotah-prod/r/steamwheedle-cartel.git/pkg/logging"
)

// Work - spawns N number of goroutines to execute X() in parallel, with Y() called when they exit
func Work(workerCount int, worker func(), postWork func()) {
	wg := &sync.WaitGroup{}
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			worker()
		}()
	}

	go func() {
		wg.Wait()
		postWork()
	}()
}

// Download - performs HTTP GET request against url, including adding gzip header and ungzipping
func Download(url string) (b []byte, err error) {
	var (
		req    *http.Request
		reader io.ReadCloser
	)

	// forming a request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return b, err
	}
	req.Header.Add("Accept-Encoding", "gzip")

	// running it into a client
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return b, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logging.WithField("error", err.Error()).Error("failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return b, fmt.Errorf("response was not OK: %d", resp.StatusCode)
	}

	// optionally decompressing it
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return
		}
		defer func() {
			if err := reader.Close(); err != nil {
				logging.WithField("error", err.Error()).Error("failed to close reader body")
			}
		}()
	default:
		reader = resp.Body
	}

	return ioutil.ReadAll(reader)
}

// ReadFile - reads a file from a relative path
func ReadFile(relativePath string) ([]byte, error) {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		return []byte{}, err
	}

	return ioutil.ReadFile(path)
}

// WriteFile - writes a file using a relative path
func WriteFile(relativePath string, data []byte) error {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

// GzipEncode - gzip encodes a byte array
func GzipEncode(in []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write(in); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// GzipDecode - gzip decodes a byte array
func GzipDecode(in []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := r.Close(); err != nil {
			logging.WithField("error", err.Error()).Error("failed to close reader")
		}
	}()

	return ioutil.ReadAll(r)
}

// EnsureDirExists - ensures dir exists
func EnsureDirExists(relativePath string) error {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

// EnsureDirsExist - ensuring dirs exist
func EnsureDirsExist(relativePaths []string) error {
	for _, relativePath := range relativePaths {
		if err := EnsureDirExists(relativePath); err != nil {
			return err
		}
	}

	return nil
}
