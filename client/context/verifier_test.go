package context_test

import (
	"io/ioutil"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/stretchr/testify/require"
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
		t.Run(tc.name, func(t *testing.T) {
			verifier, err := context.CreateVerifier(tc.ctx, context.DefaultVerifierCacheSize)
			require.Equal(t, tc.expectErr, err != nil, err)

			if !tc.expectErr {
				require.NotNil(t, verifier)
			}
		})
	}
}
