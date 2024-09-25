package systemtests

import (
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

type GRPCTestCaseWithHeaders struct {
	name    string
	url     string
	headers map[string]string
	expOut  string
}

// RunGRPCQueriesWithHeaders runs given grpc testcases by making requests with headers and
// checking response with expected output
func RunGRPCQueriesWithHeaders(t *testing.T, testCases []GRPCTestCaseWithHeaders) {
	t.Helper()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := testutil.GetRequestWithHeaders(tc.url, tc.headers)
			require.NoError(t, err)
			require.JSONEq(t, tc.expOut, string(resp))
		})
	}
}
