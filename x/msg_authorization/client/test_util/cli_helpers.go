package test_util

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	msgauthcli "github.com/cosmos/cosmos-sdk/x/msg_authorization/client/cli"
)

func MsgGrantSendAuthorizationExec(clientCtx client.Context, grantee fmt.Stringer, msgType, authorization string, extraArgs ...string) (testutil.BufferWriter, error) {
	args := []string{grantee.String(), msgType, authorization}
	args = append(args, extraArgs...)

	return clitestutil.ExecTestCLICmd(clientCtx, msgauthcli.GetCmdGrantAuthorization(""), args)
}
