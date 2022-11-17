package types_test

import (
	"testing"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
)

func TestNewModuleCrendentials(t *testing.T) {
	expected := sdk.MustAccAddressFromBech32("cosmos1fpn0w0yf4x300llf5r66jnfhgj4ul6cfahrvqsskwkhsw6sv84wsmz359y")

	credential := authtypes.NewModuleCredential("group", [][]byte{{0x20}, {0x0}})
	require.NoError(t, sdk.VerifyAddressFormat(credential.Address().Bytes()))
	addr, err := sdk.AccAddressFromHexUnsafe(credential.Address().String())
	require.NoError(t, err)
	require.Equal(t, expected.String(), addr.String())
}

func TestNewAccountWithModuleCredential(t *testing.T) {
	expected := sdk.MustAccAddressFromBech32("cosmos1fpn0w0yf4x300llf5r66jnfhgj4ul6cfahrvqsskwkhsw6sv84wsmz359y")

	credential := authtypes.NewModuleCredential("group", [][]byte{{0x20}, {0x0}})
	account, err := authtypes.NewAccountWithModuleCredential(credential)
	require.NoError(t, err)
	require.Equal(t, expected, account.GetAddress())
	require.Equal(t, credential, account.GetPubKey())
}

func TestNewAccountWithModuleCredential_WrongCredentials(t *testing.T) {
	_, err := authtypes.NewAccountWithModuleCredential(cryptotypes.PubKey(nil))
	require.Error(t, err)
}
