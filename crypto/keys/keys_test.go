package keys_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	// "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/sr25519"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

func TestPubKeyEquals(t *testing.T) {
	sr25519PubKey := sr25519.GenPrivKey().PubKey().(sr25519.PubKey)
	sr25519PbPubKey := &keys.Sr25519PubKey{Key: sr25519PubKey}

	ed25519PubKey := ed25519.GenPrivKey().PubKey().(ed25519.PubKey)
	ed25519PbPubKey := &keys.Ed25519PubKey{Key: ed25519PubKey}

	secp256K1PubKey := secp256k1.GenPrivKey().PubKey().(secp256k1.PubKey)
	secp256K1PbPubKey := &keys.Secp256K1PubKey{Key: secp256K1PubKey}

	testCases := []struct {
		msg      string
		pubKey   crypto.PubKey
		other    crypto.PubKey
		expectEq bool
	}{
		{
			"sr25519 pb different types",
			sr25519PbPubKey,
			&keys.Ed25519PubKey{
				Key: ed25519.GenPrivKey().PubKey().(ed25519.PubKey),
			},
			false,
		},
		{
			"sr25519 pb different bytes",
			sr25519PbPubKey,
			&keys.Sr25519PubKey{
				Key: sr25519.GenPrivKey().PubKey().(sr25519.PubKey),
			},
			false,
		},
		{
			"sr25519 pb equals",
			sr25519PbPubKey,
			&keys.Sr25519PubKey{
				Key: sr25519PubKey,
			},
			true,
		},
		{
			"sr25519 different types",
			sr25519PbPubKey,
			ed25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"sr25519 different bytes",
			sr25519PbPubKey,
			sr25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"sr25519 equals",
			sr25519PbPubKey,
			sr25519PubKey,
			true,
		},
		{
			"ed25519 pb different types",
			ed25519PbPubKey,
			&keys.Sr25519PubKey{
				Key: sr25519.GenPrivKey().PubKey().(sr25519.PubKey),
			},
			false,
		},
		{
			"ed25519 pb different bytes",
			ed25519PbPubKey,
			&keys.Ed25519PubKey{
				Key: ed25519.GenPrivKey().PubKey().(ed25519.PubKey),
			},
			false,
		},
		{
			"ed25519 pb equals",
			ed25519PbPubKey,
			&keys.Ed25519PubKey{
				Key: ed25519PubKey,
			},
			true,
		},
		{
			"tm ed25519 equals pb ed25519",
			ed25519PubKey,
			&keys.Ed25519PubKey{
				Key: ed25519PubKey,
			},
			true,
		},
		{
			"ed25519 different types",
			ed25519PbPubKey,
			sr25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"ed25519 different bytes",
			ed25519PbPubKey,
			ed25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"ed25519 equals",
			ed25519PbPubKey,
			ed25519PubKey,
			true,
		},
		{
			"secp256k1 pb different types",
			secp256K1PbPubKey,
			&keys.Sr25519PubKey{
				Key: sr25519.GenPrivKey().PubKey().(sr25519.PubKey),
			},
			false,
		},
		{
			"secp256k1 pb different bytes",
			secp256K1PbPubKey,
			&keys.Secp256K1PubKey{
				Key: secp256k1.GenPrivKey().PubKey().(secp256k1.PubKey),
			},
			false,
		},
		{
			"secp256k1 pb equals",
			secp256K1PbPubKey,
			&keys.Secp256K1PubKey{
				Key: secp256K1PubKey,
			},
			true,
		},
		{
			"secp256k1 different types",
			secp256K1PbPubKey,
			sr25519.GenPrivKey().PubKey(),
			false,
		},
		{
			"secp256k1 different bytes",
			secp256K1PbPubKey,
			secp256k1.GenPrivKey().PubKey(),
			false,
		},
		{
			"secp256k1 equals",
			secp256K1PbPubKey,
			secp256K1PubKey,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.pubKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestPrivKeyEquals(t *testing.T) {
	sr25519PrivKey := sr25519.GenPrivKey()
	sr25519PbPrivKey := &keys.Sr25519PrivKey{Key: sr25519PrivKey}

	ed25519PrivKey := ed25519.GenPrivKey()
	ed25519PbPrivKey := &keys.Ed25519PrivKey{Key: ed25519PrivKey}

	secp256K1PrivKey := secp256k1.GenPrivKey()
	secp256K1PbPrivKey := &keys.Secp256K1PrivKey{Key: secp256K1PrivKey}

	testCases := []struct {
		msg      string
		privKey  crypto.PrivKey
		other    crypto.PrivKey
		expectEq bool
	}{
		{
			"sr25519 pb different types",
			sr25519PbPrivKey,
			&keys.Ed25519PrivKey{
				Key: ed25519.GenPrivKey(),
			},
			false,
		},
		{
			"sr25519 pb different bytes",
			sr25519PbPrivKey,
			&keys.Sr25519PrivKey{
				Key: sr25519.GenPrivKey(),
			},
			false,
		},
		{
			"sr25519 pb equals",
			sr25519PbPrivKey,
			&keys.Sr25519PrivKey{
				Key: sr25519PrivKey,
			},
			true,
		},
		{
			"sr25519 different types",
			sr25519PbPrivKey,
			ed25519.GenPrivKey(),
			false,
		},
		{
			"sr25519 different bytes",
			sr25519PbPrivKey,
			sr25519.GenPrivKey(),
			false,
		},
		{
			"sr25519 equals",
			sr25519PbPrivKey,
			sr25519PrivKey,
			true,
		},
		{
			"ed25519 pb different types",
			ed25519PbPrivKey,
			&keys.Sr25519PrivKey{
				Key: sr25519.GenPrivKey(),
			},
			false,
		},
		{
			"ed25519 pb different bytes",
			ed25519PbPrivKey,
			&keys.Ed25519PrivKey{
				Key: ed25519.GenPrivKey(),
			},
			false,
		},
		{
			"ed25519 pb equals",
			ed25519PbPrivKey,
			&keys.Ed25519PrivKey{
				Key: ed25519PrivKey,
			},
			true,
		},
		{
			"ed25519 different types",
			ed25519PbPrivKey,
			sr25519.GenPrivKey(),
			false,
		},
		{
			"ed25519 different bytes",
			ed25519PbPrivKey,
			ed25519.GenPrivKey(),
			false,
		},
		{
			"ed25519 equals",
			ed25519PbPrivKey,
			ed25519PrivKey,
			true,
		},
		{
			"secp256k1 pb different types",
			secp256K1PbPrivKey,
			&keys.Sr25519PrivKey{
				Key: sr25519.GenPrivKey(),
			},
			false,
		},
		{
			"secp256k1 pb different bytes",
			secp256K1PbPrivKey,
			&keys.Secp256K1PrivKey{
				Key: secp256k1.GenPrivKey(),
			},
			false,
		},
		{
			"secp256k1 pb equals",
			secp256K1PbPrivKey,
			&keys.Secp256K1PrivKey{
				Key: secp256K1PrivKey,
			},
			true,
		},
		{
			"secp256k1 different types",
			secp256K1PbPrivKey,
			sr25519.GenPrivKey(),
			false,
		},
		{
			"secp256k1 different bytes",
			secp256K1PbPrivKey,
			secp256k1.GenPrivKey(),
			false,
		},
		{
			"secp256k1 equals",
			secp256K1PbPrivKey,
			secp256K1PrivKey,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.privKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}
