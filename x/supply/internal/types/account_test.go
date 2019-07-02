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
	moduleAcc := NewEmptyModuleAccount(name, Basic)
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
  permission: %s
`, moduleAddress, name, Basic)

	require.Equal(t, want, string(bs))
	require.Equal(t, want, moduleAcc.String())
}
