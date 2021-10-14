// Package rest provides HTTP types and primitives for REST
// requests validation and responses handling.
package rest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// GetRequest defines a wrapper around an HTTP GET request with a provided URL.
// An error is returned if the request or reading the body fails.
func GetRequest(url string) ([]byte, error) {
	res, err := http.Get(url) // nolint:gosec
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// PostRequest defines a wrapper around an HTTP POST request with a provided URL and data.
// An error is returned if the request or reading the body fails.
func PostRequest(url string, contentType string, data []byte) ([]byte, error) {
	res, err := http.Post(url, contentType, bytes.NewBuffer(data)) // nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("error while sending post request: %w", err)
	}
	defer res.Body.Close()

	bz, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return bz, nil
}
