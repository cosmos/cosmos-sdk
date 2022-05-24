package testutil

import (
	"github.com/Stride-Labs/cosmos-sdk/testutil"
	clitestutil "github.com/Stride-Labs/cosmos-sdk/testutil/cli"
	"github.com/Stride-Labs/cosmos-sdk/testutil/network"
	"github.com/Stride-Labs/cosmos-sdk/x/authz/client/cli"
)

func ExecGrant(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := cli.NewCmdGrantAuthorization()
	clientCtx := val.ClientCtx
	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}
