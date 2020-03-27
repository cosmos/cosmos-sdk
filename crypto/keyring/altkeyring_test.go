package keyring

import (
	"testing"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
)

func TestAltKeyring_List(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)

	// Create 3 keys
	uid1, uid2, uid3 := "Zkey", "Bkey", "Rkey"
	_, _, err = keyring.NewMnemonic(uid1, English, Secp256k1)
	require.NoError(t, err)
	_, _, err = keyring.NewMnemonic(uid2, English, Secp256k1)
	require.NoError(t, err)
	_, _, err = keyring.NewMnemonic(uid3, English, Secp256k1)
	require.NoError(t, err)

	list, err = keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 3)

	// Check they are in alphabetical order
	require.Equal(t, uid2, list[0].GetName())
	require.Equal(t, uid3, list[1].GetName())
	require.Equal(t, uid1, list[2].GetName())
}

func TestAltKeyring_Get(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theKey"
	mnemonic, _, err := keyring.NewMnemonic(uid, English, Secp256k1)
	require.NoError(t, err)

	key, err := keyring.Key(uid)
	require.NoError(t, err)
	requireEqualInfo(t, mnemonic, key)
}

func TestAltKeyring_KeyByAddress(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theKey"
	mnemonic, _, err := keyring.NewMnemonic(uid, English, Secp256k1)
	require.NoError(t, err)

	key, err := keyring.KeyByAddress(mnemonic.GetAddress())
	require.NoError(t, err)
	requireEqualInfo(t, key, mnemonic)
}

func TestAltKeyring_Delete(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theKey"
	_, _, err = keyring.NewMnemonic(uid, English, Secp256k1)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 1)

	err = keyring.Delete(uid)
	require.NoError(t, err)

	list, err = keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestAltKeyring_DeleteByAddress(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	uid := "theKey"
	mnemonic, _, err := keyring.NewMnemonic(uid, English, Secp256k1)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 1)

	err = keyring.DeleteByAddress(mnemonic.GetAddress())
	require.NoError(t, err)

	list, err = keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestAltKeyring_SavePubKey(t *testing.T) {
	dir, clean := tests.NewTestCaseDir(t)
	t.Cleanup(clean)

	keyring, err := NewAltKeyring(t.Name(), BackendTest, dir, nil)
	require.NoError(t, err)

	list, err := keyring.List()
	require.NoError(t, err)
	require.Empty(t, list)

	key := "offline"
	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()

	info, err := keyring.SavePubKey(key, pub, Ed25519)
	require.Nil(t, err)
	require.Equal(t, pub, info.GetPubKey())
	require.Equal(t, key, info.GetName())

	list, err = keyring.List()
	require.NoError(t, err)
	require.Equal(t, 1, len(list))
}

func requireEqualInfo(t *testing.T, key Info, mnemonic Info) {
	require.Equal(t, key.GetName(), mnemonic.GetName())
	require.Equal(t, key.GetAddress(), mnemonic.GetAddress())
	require.Equal(t, key.GetPubKey(), mnemonic.GetPubKey())
	require.Equal(t, key.GetAlgo(), mnemonic.GetAlgo())
	require.Equal(t, key.GetType(), mnemonic.GetType())
}
