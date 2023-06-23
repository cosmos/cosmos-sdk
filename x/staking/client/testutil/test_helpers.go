package testutil

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
}

func ExecuteCmdAndCheckSuccess(t *testing.T, clientContext client.Context, cmd *cobra.Command, extraArgs []string) {
	out, err := clitestutil.ExecTestCLICmd(clientContext, cmd, extraArgs)
	require.NoError(t, err)

	var txRes sdk.TxResponse
	clientContext.Codec.UnmarshalJSON(out.Bytes(), &txRes)
	require.Equal(t, int64(0), int64(txRes.Code), "Response returned non-zero exit code:\n%+v", txRes)
}

// MsgRedelegateExec creates a redelegate message.
func MsgRedelegateExec(
	t *testing.T, clientCtx client.Context,
	from, src, dst, amount fmt.Stringer,
	extraArgs ...string,
) {
	args := []string{
		src.String(),
		dst.String(),
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from.String()),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 300000),
	}
	args = append(args, extraArgs...)

	args = append(args, commonArgs...)
	ExecuteCmdAndCheckSuccess(t, clientCtx, stakingcli.NewRedelegateCmd(), args)
}

// MsgUnbondExec creates a unbond message.
func MsgUnbondExec(
	t *testing.T, clientCtx client.Context,
	from fmt.Stringer, valAddress, amount fmt.Stringer,
	extraArgs ...string,
) {
	args := []string{
		valAddress.String(),
		amount.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from.String()),
	}

	args = append(args, commonArgs...)
	args = append(args, extraArgs...)
	ExecuteCmdAndCheckSuccess(t, clientCtx, stakingcli.NewUnbondCmd(), args)
}

// MsgTokenizeSharesExec creates a delegation message.
func MsgTokenizeSharesExec(
	t *testing.T, clientCtx client.Context,
	from fmt.Stringer, valAddress, rewardOwner, amount fmt.Stringer,
	extraArgs ...string,
) {
	args := []string{
		valAddress.String(),
		amount.String(),
		rewardOwner.String(),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from.String()),
	}

	args = append(args, commonArgs...)
	args = append(args, extraArgs...)
	ExecuteCmdAndCheckSuccess(t, clientCtx, stakingcli.NewTokenizeSharesCmd(), args)
}
