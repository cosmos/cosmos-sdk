package systemtests

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
)

type GRPCTestCase struct {
	name   string
	url    string
	expOut string
}

// RunGRPCQueries runs given grpc testcases by making requests and
// checking response with expected output
func RunGRPCQueries(t *testing.T, testCases []GRPCTestCase) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequest(tc.url)
			require.NoError(t, err)
			require.JSONEq(t, tc.expOut, string(resp))
		})
	}
}

// TestGRPCQueryIgnoreNumbers runs given grpc testcases by making requests and
// checking response with expected output ignoring number values
// This method is using when number values in response are non-deterministic
func TestGRPCQueryIgnoreNumbers(t *testing.T, testCases []GRPCTestCase) {
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
