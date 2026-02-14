package systemtests

import (
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

type RestTestCase struct {
	Name       string
	Url        string
	ExpCode    int
	ExpCodeGTE int
	ExpOut     string
}

// RunRestQueries runs given Rest testcases by making requests and
// checking response with expected output
func RunRestQueries(t *testing.T, testCases ...RestTestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.ExpCodeGTE > 0 && tc.ExpCode > 0 {
				require.Fail(t, "only one of ExpCode or ExpCodeGTE should be set")
			}

			var resp []byte
			if tc.ExpCodeGTE > 0 {
				resp = GetRequestWithHeadersGreaterThanOrEqual(t, tc.Url, nil, tc.ExpCodeGTE)
			} else {
				resp = GetRequestWithHeaders(t, tc.Url, nil, tc.ExpCode)
			}
			require.JSONEq(t, tc.ExpOut, string(resp))
		})
	}
}

// RunRestQueriesIgnoreNumbers runs given rest testcases by making requests and
// checking response with expected output ignoring number values
// This method is used when number values in response are non-deterministic
func RunRestQueriesIgnoreNumbers(t *testing.T, testCases ...RestTestCase) {
	t.Helper()

	// regex for standalone quoted numbers (e.g., "-3" or "0.02")
	standaloneQuotedNumberRegex := regexp.MustCompile(`"(-?\d+(\.\d+)?)"`)
	// regex for numbers in escaped strings (e.g., \"-3\")
	escapedNumberRegex := regexp.MustCompile(`\\\"(-?\d+(\.\d+)?)\\\"`)
	// regex for unquoted numbers (e.g., 2, -1, 0.02,)
	unquotedNumberRegex := regexp.MustCompile(`\b-?\d+(\.\d+)?\b,`)

	replaceNumber := func(input string) string {
		// handle numbers in escaped strings
		result := escapedNumberRegex.ReplaceAllStringFunc(input, func(match string) string {
			// replace with escaped "NUMBER"
			return `\"NUMBER\"`
		})

		// handle standalone quoted numbers
		result = standaloneQuotedNumberRegex.ReplaceAllStringFunc(result, func(match string) string {
			// replace with "NUMBER" (quotes preserved)
			return `"NUMBER"`
		})

		// handle unquoted numbers
		result = unquotedNumberRegex.ReplaceAllStringFunc(result, func(match string) string {
			// replace with "NUMBER" (add quotes to ensure json validity)
			return `"NUMBER",`
		})

		return result
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.ExpCodeGTE > 0 && tc.ExpCode > 0 {
				require.Fail(t, "only one of ExpCode or ExpCodeGTE should be set")
			}

			var resp []byte
			if tc.ExpCodeGTE > 0 {
				resp = GetRequestWithHeadersGreaterThanOrEqual(t, tc.Url, nil, tc.ExpCodeGTE)
			} else {
				resp = GetRequestWithHeaders(t, tc.Url, nil, tc.ExpCode)
			}

			expectedJSON := replaceNumber(tc.ExpOut)
			actualJSON := replaceNumber(string(resp))

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

func GetRequestWithHeadersGreaterThanOrEqual(t *testing.T, url string, headers map[string]string, expCode int) []byte {
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
	require.GreaterOrEqual(t, res.StatusCode, expCode, "status code should be greater or equal to %d, got: %d, %s", expCode, res.StatusCode, body)

	return body
}
