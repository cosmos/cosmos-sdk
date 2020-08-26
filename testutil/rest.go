package testutil

import (
	"io/ioutil"
	"net/http"
)

// GetRequestWithHeaders defines a wrapper around an HTTP GET request with a provided URL
// and custom headers
// An error is returned if the request or reading the body fails.
func GetRequestWithHeaders(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if err = res.Body.Close(); err != nil {
		return nil, err
	}

	return body, nil
}
