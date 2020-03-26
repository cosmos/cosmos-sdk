package keyring

import (
	"testing"

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
