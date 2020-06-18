package client_test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
)

func TestCreateVerifier(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "example")
	require.NoError(t, err)

	testCases := []struct {
		name      string
		ctx       client.Context
		expectErr bool
	}{
		{"no chain ID", client.Context{}, true},
		{"no home directory", client.Context{}.WithChainID("test"), true},
		{"no client or RPC URI", client.Context{HomeDir: tmpDir}.WithChainID("test"), true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			verifier, err := client.CreateVerifier(tc.ctx, client.DefaultVerifierCacheSize)
			require.Equal(t, tc.expectErr, err != nil, err)

			if !tc.expectErr {
				require.NotNil(t, verifier)
			}
		})
	}
}
