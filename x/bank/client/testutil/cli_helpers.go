package testutil

import (
	"bytes"
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

func MsgSendExec(clientCtx client.Context, from, to, amount fmt.Stringer, extraArgs ...string) ([]byte, error) {
	args := []string{from.String(), to.String(), amount.String()}
	args = append(args, extraArgs...)

	return callCmd(clientCtx, bankcli.NewSendTxCmd, args)
}

func QueryBalancesExec(clientCtx client.Context, address sdk.AccAddress, extraArgs ...string) ([]byte, error) {
	args := []string{address.String()}
	args = append(args, extraArgs...)

	return callCmd(clientCtx, bankcli.GetBalancesCmd, args)
}

func callCmd(clientCtx client.Context, theCmd func() *cobra.Command, extraArgs []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	clientCtx = clientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := theCmd()
	cmd.SetErr(buf)
	cmd.SetOut(buf)

	cmd.SetArgs(extraArgs)

	if err := cmd.ExecuteContext(ctx); err != nil {
		return buf.Bytes(), err
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

// QueryBalances executes the bank query balances command for a given address and
// flag set.
func QueryBalances(f *cli.Fixtures, address sdk.AccAddress, flags ...string) sdk.Coins {
	cmd := fmt.Sprintf("%s query bank balances %s %v", f.SimdBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")

	var balances sdk.Coins

	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &balances), "out %v\n", out)

	return balances
}
