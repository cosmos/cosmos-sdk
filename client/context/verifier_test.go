package context_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/context"
)

func TestCreateVerifier(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "example")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		ctx       context.CLIContext
		expectErr bool
	}{
		{"no chain ID", context.CLIContext{}, true},
		{"no home directory", context.CLIContext{}.WithChainID("test"), true},
		{"no client or RPC URI", context.CLIContext{HomeDir: tmpDir}.WithChainID("test"), true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			verifier, err := context.CreateVerifier(tc.ctx, context.DefaultVerifierCacheSize)
			require.Equal(t, tc.expectErr, err != nil, err)

			if !tc.expectErr {
				require.NotNil(t, verifier)
			}
		})
	}
}
