package context

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateVerifier(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "example")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		ctx       CLIContext
		expectErr bool
	}{
		{"no chain ID", CLIContext{}, true},
		{"no home directory", CLIContext{}.WithChainID("test"), true},
		{"no client or RPC URI", CLIContext{HomeDir: tmpDir}.WithChainID("test"), true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			verifier, err := CreateVerifier(tc.ctx, DefaultVerifierCacheSize)
			require.Equal(t, tc.expectErr, err != nil, err)

			if !tc.expectErr {
				require.NotNil(t, verifier)
			}
		})
	}
}
