package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types"
)

func TestGenerateCoinKey(t *testing.T) {
	t.Parallel()
	addr, mnemonic, err := server.GenerateCoinKey(hd.Secp256k1)
	require.NoError(t, err)

	// Test creation
	info, err := keyring.NewInMemory().NewAccount("xxx", mnemonic, "", hd.NewFundraiserParams(0, types.GetConfig().GetCoinType(), 0).String(), hd.Secp256k1)
	require.NoError(t, err)
	require.Equal(t, addr, info.GetAddress())
}

func TestGenerateSaveCoinKey(t *testing.T) {
	t.Parallel()

	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil)
	require.NoError(t, err)

	addr, mnemonic, err := server.GenerateSaveCoinKey(kb, "keyname", false, hd.Secp256k1)
	require.NoError(t, err)

	// Test key was actually saved
	info, err := kb.Key("keyname")
	require.NoError(t, err)
	require.Equal(t, addr, info.GetAddress())

	// Test in-memory recovery
	info, err = keyring.NewInMemory().NewAccount("xxx", mnemonic, "", hd.NewFundraiserParams(0, types.GetConfig().GetCoinType(), 0).String(), hd.Secp256k1)
	require.NoError(t, err)
	require.Equal(t, addr, info.GetAddress())
}

func TestGenerateSaveCoinKeyOverwriteFlag(t *testing.T) {
	t.Parallel()

	kb, err := keyring.New(t.Name(), "test", t.TempDir(), nil)
	require.NoError(t, err)

	keyname := "justakey"
	addr1, _, err := server.GenerateSaveCoinKey(kb, keyname, false, hd.Secp256k1)
	require.NoError(t, err)

	// Test overwrite with overwrite=false
	_, _, err = server.GenerateSaveCoinKey(kb, keyname, false, hd.Secp256k1)
	require.Error(t, err)

	// Test overwrite with overwrite=true
	addr2, _, err := server.GenerateSaveCoinKey(kb, keyname, true, hd.Secp256k1)
	require.NoError(t, err)

	require.NotEqual(t, addr1, addr2)
}
