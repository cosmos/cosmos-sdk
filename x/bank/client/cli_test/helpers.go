package cli

import (
	"encoding/json"
	"fmt"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
)

// TxSend is simcli tx send
func TxSend(f *helpers.Fixtures, from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send --keyring-backend=test %s %s %s %v", f.SimcliBinary, from,
		to, amount, f.Flags())
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryAccount is simcli query account
func QueryAccount(f *helpers.Fixtures, address sdk.AccAddress, flags ...string) auth.BaseAccount {
	cmd := fmt.Sprintf("%s query account %s %v", f.SimcliBinary, address, f.Flags())

	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)
	value := initRes["value"]

	var acc auth.BaseAccount
	err = f.Cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)

	return acc
}

// QueryBalances executes the bank query balances command for a given address and
// flag set.
func QueryBalances(f *helpers.Fixtures, address sdk.AccAddress, flags ...string) sdk.Coins {
	cmd := fmt.Sprintf("%s query bank balances %s %v", f.SimcliBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")

	var balances sdk.Coins

	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &balances), "out %v\n", out)

	return balances
}
