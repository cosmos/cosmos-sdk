package hd_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
)

const (
	testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	hdPath0      = "m/44'/118'/0'/0/0"
	hdPath1      = "m/44'/118'/0'/0/1"
)

func TestMlDsa65Name(t *testing.T) {
	require.Equal(t, hd.MlDsa65Type, hd.MlDsa65.Name())
}

func TestMlDsa65DeriveAndGenerate(t *testing.T) {
	derive := hd.MlDsa65.Derive()
	gen := hd.MlDsa65.Generate()

	seed, err := derive(testMnemonic, "", hdPath0)
	require.NoError(t, err)
	require.Len(t, seed, 32) // matches mldsa65 SeedSize

	priv := gen(seed)
	require.NotNil(t, priv)
	_, ok := priv.(*mldsa65.PrivKey)
	require.True(t, ok)
}

func TestMlDsa65GenerateZeroSeedNoPanic(t *testing.T) {
	require.NotPanics(t, func() {
		priv := hd.MlDsa65.Generate()(make([]byte, 32))
		require.NotNil(t, priv)
		require.Equal(t, "ml_dsa_65", priv.PubKey().Type())
	})
}

func TestMlDsa65DerivationDeterministic(t *testing.T) {
	derive := hd.MlDsa65.Derive()
	gen := hd.MlDsa65.Generate()

	mk := func(path string) []byte {
		seed, err := derive(testMnemonic, "", path)
		require.NoError(t, err)
		return gen(seed).PubKey().Bytes()
	}

	// Same mnemonic+path -> same key (recovery).
	require.Equal(t, mk(hdPath0), mk(hdPath0))
	// Different path -> different account (path salt).
	require.NotEqual(t, mk(hdPath0), mk(hdPath1))
}
