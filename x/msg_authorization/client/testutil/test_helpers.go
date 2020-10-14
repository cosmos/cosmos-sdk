package testutil

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/x/msg_authorization/client/cli"
	msgauthcli "github.com/cosmos/cosmos-sdk/x/msg_authorization/client/cli"
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
		limit,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, granter),
	}

	viper.Set(cli.FlagExpiration, time.Now().Add(time.Minute*time.Duration(120)).Unix())

	args = append(args, commonArgs...)
	return clitestutil.ExecTestCLICmd(clientCtx, msgauthcli.NewCmdGrantAuthorization(), args)

}
