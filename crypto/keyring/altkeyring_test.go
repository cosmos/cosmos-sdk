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
	require.Len(t, list, 0)

	_, _, err = keyring.NewMnemonic("theKey", English, Secp256k1)
	require.NoError(t, err)

	list, err = keyring.List()
	require.NoError(t, err)
	require.Len(t, list, 1)
}
