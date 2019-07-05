package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/require"
)

func TestModuleAccountMarshalYAML(t *testing.T) {
	name := "test"
	moduleAcc := NewEmptyModuleAccount(name, Basic, Minter, Burner)
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	bs, err := yaml.Marshal(moduleAcc)
	require.NoError(t, err)

	want := fmt.Sprintf(`|
  address: %s
  coins: []
  pubkey: ""
  accountnumber: 0
  sequence: 0
  name: %s
  permissions:
  - %s
  - %s
  - %s
`, moduleAddress, name, Basic, Minter, Burner)

	require.Equal(t, want, string(bs))
	require.Equal(t, want, moduleAcc.String())
}

func TestRemovePermissions(t *testing.T) {
	name := "test"
	macc := NewEmptyModuleAccount(name)
	require.Empty(t, macc.GetPermissions())

	macc.AddPermissions(Basic, Minter, Burner)
	require.Equal(t, []string{Basic, Minter, Burner}, macc.GetPermissions(), "did not add permissions")

	err := macc.RemovePermission("random")
	require.Error(t, err, "did not error on removing nonexistent permission")

	err = macc.RemovePermission(Burner)
	require.NoError(t, err, "failed to remove permission")
	require.Equal(t, []string{Basic, Minter}, macc.GetPermissions(), "does not have correct permissions")

	err = macc.RemovePermission(Basic)
	require.NoError(t, err, "failed to remove permission")
	require.Equal(t, []string{Minter}, macc.GetPermissions(), "does not have correct permissions")
}

func TestHasPermissions(t *testing.T) {
	name := "test"
	macc := NewEmptyModuleAccount(name, Basic, Minter, Burner)
	cases := []struct {
		permission string
		expectHas  bool
	}{
		{Basic, true},
		{Minter, true},
		{Burner, true},
		{"other", false},
	}

	for i, tc := range cases {
		hasPerm := macc.HasPermission(tc.permission)
		if tc.expectHas {
			require.True(t, hasPerm, "test case #%d", i)
		} else {
			require.False(t, hasPerm, "test case #%d", i)
		}
	}
}
