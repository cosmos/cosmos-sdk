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
	moduleAcc := NewEmptyModuleAccount(name, Minter, Burner, Staking)
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
`, moduleAddress, name, Minter, Burner, Staking)

	require.Equal(t, want, string(bs))
	require.Equal(t, want, moduleAcc.String())
}

func TestHasPermissions(t *testing.T) {
	name := "test"
	macc := NewEmptyModuleAccount(name, Staking, Minter, Burner)
	cases := []struct {
		permission string
		expectHas  bool
	}{
		{Staking, true},
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
