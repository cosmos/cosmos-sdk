package systemtests

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type RestTestCase struct {
	name    string
	url     string
	expCode int
	expOut  string
}

// RunRestQueries runs given Rest testcases by making requests and
// checking response with expected output
func RunRestQueries(t *testing.T, testCases []RestTestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := GetRequestWithHeaders(t, tc.url, nil, tc.expCode)
			require.JSONEq(t, tc.expOut, string(resp))
		})
	}
}

func GetRequest(t *testing.T, url string) []byte {
	t.Helper()
	return GetRequestWithHeaders(t, url, nil, http.StatusOK)
}

func GetRequestWithHeaders(t *testing.T, url string, headers map[string]string, expCode int) []byte {
	t.Helper()
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	httpClient := &http.Client{}
	res, err := httpClient.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = res.Body.Close()
	}()
	require.Equal(t, expCode, res.StatusCode, "status code should be %d, got: %d", expCode, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return body
}
