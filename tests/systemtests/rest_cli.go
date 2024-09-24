package systemtests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/stretchr/testify/require"
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
