package server_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/keys"
	keys2 "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types"
)

func TestGenerateCoinKey(t *testing.T) {
	t.Parallel()
	addr, mnemonic, err := server.GenerateCoinKey()
	require.NoError(t, err)

	// Test creation
	info, err := keys.NewInMemoryKeyBase().CreateAccount("xxx", mnemonic, "", "012345678", 0, 0)
	require.NoError(t, err)
	require.Equal(t, addr, info.GetAddress())
}

func TestGenerateSaveCoinKey(t *testing.T) {
	t.Parallel()
	dir, cleanup := tempdir(t)
	defer cleanup() // clean after itself
	// Remove the dir to that GenerateSaveCoinKey creates it automatically
	os.RemoveAll(dir)

	addr, mnemonic, err := server.GenerateSaveCoinKey(dir, "keyname", "012345678", false, nil)
	require.NoError(t, err)

	// Test key was actually saved
	kb, err := keys2.NewTestKeyring(types.GetConfig().GetKeyringServiceName(), dir)
	require.NoError(t, err)
	info, err := kb.Get("keyname")
	require.NoError(t, err)
	require.Equal(t, addr, info.GetAddress())

	// Test in-memory recovery
	info, err = keys.NewInMemoryKeyBase().CreateAccount("xxx", mnemonic, "", "012345678", 0, 0)
	require.NoError(t, err)
	require.Equal(t, addr, info.GetAddress())
}

func TestGenerateSaveCoinKeyOverwriteFlag(t *testing.T) {
	t.Parallel()
	dir, cleanup := tempdir(t)
	defer cleanup() // clean after itself
	// Remove the dir to that GenerateSaveCoinKey creates it automatically
	os.RemoveAll(dir)

	keyname := "justakey"
	addr1, _, err := server.GenerateSaveCoinKey(dir, keyname, "012345678", false, nil)
	require.NoError(t, err)

	// Test overwrite with overwrite=false
	_, _, err = server.GenerateSaveCoinKey(dir, keyname, "012345678", false, nil)
	require.Error(t, err)

	// Test overwrite with overwrite=true
	addr2, _, err := server.GenerateSaveCoinKey(dir, keyname, "012345678", true, nil)
	require.NoError(t, err)

	require.NotEqual(t, addr1, addr2)
}

func tempdir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", t.Name()+"_")
	require.NoError(t, err)
	return dir, func() { os.RemoveAll(dir) }
}
