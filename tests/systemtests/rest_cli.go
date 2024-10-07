package systemtests

import (
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
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

// TestRestQueryIgnoreNumbers runs given rest testcases by making requests and
// checking response with expected output ignoring number values
// This method is used when number values in response are non-deterministic
func TestRestQueryIgnoreNumbers(t *testing.T, testCases []RestTestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			require.NoError(t, err)

			// regular expression pattern to match any numeric value in the JSON
			numberRegexPattern := `"\d+(\.\d+)?"`

			// compile the regex
			r, err := regexp.Compile(numberRegexPattern)
			require.NoError(t, err)

			// replace all numeric values in the above JSONs with `NUMBER` text
			expectedJSON := r.ReplaceAllString(tc.expOut, `"NUMBER"`)
			actualJSON := r.ReplaceAllString(string(resp), `"NUMBER"`)

			// compare two jsons
			require.JSONEq(t, expectedJSON, actualJSON)
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
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, expCode, res.StatusCode, "status code should be %d, got: %d, %s", expCode, res.StatusCode, body)

	return body
}
