package testutil

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzcli "github.com/cosmos/cosmos-sdk/x/authz/client/cli"
)

var commonArgs = []string{
	fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
	fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
}

func MsgGrantAuthorizationExec(clientCtx client.Context, granter, grantee, msgName, limit string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{
		grantee,
		msgName,
	}
	if limit != "" {
		args = append(args, limit)
	}

	args = append(args, fmt.Sprintf("--%s=%s", flags.FlagFrom, granter))
	args = append(args, fmt.Sprintf("--%s=%d", authzcli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix()))
	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, authzcli.NewCmdGrantAuthorization(), args)
}
