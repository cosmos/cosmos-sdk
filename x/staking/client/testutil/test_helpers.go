package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
}

// MsgRedelegateExec creates a redelegate message.
func MsgRedelegateExec(clientCtx client.Context, from, src, dst, amount fmt.Stringer,
	extraArgs ...string) (testutil.BufferWriter, error) {

	args := []string{
		src.String(),
		dst.String(),
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from.String()),
	}

	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, stakingcli.NewRedelegateCmd(), args)
}

// MsgUnbondExec creates a unbond message.
func MsgUnbondExec(clientCtx client.Context, from fmt.Stringer, valAddress,
	amount fmt.Stringer, extraArgs ...string) (testutil.BufferWriter, error) {

	args := []string{
		valAddress.String(),
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from.String()),
	}

	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, stakingcli.NewUnbondCmd(), args)
}
