package keys_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/sr25519"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
)

func TestSr25519Equals(t *testing.T) {
	pubKey := sr25519.GenPrivKey().PubKey().(sr25519.PubKey)
	pbPubKey := &keys.Sr25519PubKey{Key: pubKey}

	testCases := []struct {
		msg      string
		other    crypto.PubKey
		expectEq bool
	}{
		{
			"different types",
			ed25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"different bytes",
			sr25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"equals",
			pubKey,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := pbPubKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}
