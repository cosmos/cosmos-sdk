package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

func MsgSendExec(clientCtx client.Context, from, to, amount fmt.Stringer, extraArgs ...string) ([]byte, error) {
	buf := new(bytes.Buffer)
	clientCtx = clientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	args := []string{from.String(), to.String(), amount.String()}
	args = append(args, extraArgs...)

	cmd := bankcli.NewSendTxCmd()
	cmd.SetErr(buf)
	cmd.SetOut(buf)
	cmd.SetArgs(args)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ----------------------------------------------------------------------------
// TODO: REMOVE ALL FUNCTIONS BELOW.
//
// REF: https://github.com/cosmos/cosmos-sdk/issues/6571
// ----------------------------------------------------------------------------

// TxSend is simcli tx send
func TxSend(f *cli.Fixtures, from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send --keyring-backend=test %s %s %s %v", f.SimdBinary, from, to, amount, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryAccount is simcli query account
func QueryAccount(f *cli.Fixtures, address sdk.AccAddress, flags ...string) authtypes.BaseAccount {
	cmd := fmt.Sprintf("%s query account %s %v", f.SimdBinary, address, f.Flags())

	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)
	value := initRes["value"]

	var acc authtypes.BaseAccount
	err = f.Cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)

	return acc
}

// QueryBalances executes the bank query balances command for a given address and
// flag set.
func QueryBalances(f *cli.Fixtures, address sdk.AccAddress, flags ...string) sdk.Coins {
	cmd := fmt.Sprintf("%s query bank balances %s %v", f.SimdBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")

	var balances sdk.Coins

	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &balances), "out %v\n", out)

	return balances
}

// QueryTotalSupply returns the total supply of coins
func QueryTotalSupply(f *cli.Fixtures, flags ...string) (totalSupply sdk.Coins) {
	cmd := fmt.Sprintf("%s query bank total %s", f.SimdBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	err := f.Cdc.UnmarshalJSON([]byte(res), &totalSupply)
	require.NoError(f.T, err)
	return totalSupply
}

// QueryTotalSupplyOf returns the total supply of a given coin denom
func QueryTotalSupplyOf(f *cli.Fixtures, denom string, flags ...string) sdk.Int {
	cmd := fmt.Sprintf("%s query bank total %s %s", f.SimdBinary, denom, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var supplyOf sdk.Int
	err := f.Cdc.UnmarshalJSON([]byte(res), &supplyOf)
	require.NoError(f.T, err)
	return supplyOf
}
