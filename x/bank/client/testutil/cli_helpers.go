package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	"github.com/cosmos/cosmos-sdk/client"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

func MsgSendExec(clientCtx client.Context, from, to, amount fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{from.String(), to.String(), amount.String()}
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, bankcli.NewSendTxCmd(), args)
}

func QueryBalancesExec(clientCtx client.Context, address fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{address.String()}
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, bankcli.GetBalancesCmd(), args)
}
