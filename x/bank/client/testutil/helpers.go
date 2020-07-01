package testutil

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

// TODO: REMOVE OR COMPLETELY REFACTOR THIS FILE.

// TxSend is simcli tx send
func TxSend(f *cli.Fixtures, from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send --keyring-backend=test %s %s %s %v", f.SimcliBinary, from, to, amount, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

func SendTx(ctx client.Context, from string, to string, amount sdk.Coins, flags ...string) ([]byte, error) {
	buf := new(bytes.Buffer)
	ctx = ctx.WithOutput(buf)

	cmd := bankcli.NewSendTxCmd(ctx)
	args := append([]string{from, to, amount.String()}, flags...)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// QueryBalances executes the bank query balances command for a given address and
// flag set.
func QueryBalances(f *cli.Fixtures, address sdk.AccAddress, flags ...string) sdk.Coins {
	cmd := fmt.Sprintf("%s query bank balances %s %v", f.SimcliBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")

	var balances sdk.Coins

	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &balances), "out %v\n", out)

	return balances
}
