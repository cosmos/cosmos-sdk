package testutil

import (
	"github.com/pointnetwork/cosmos-point-sdk/testutil"
	clitestutil "github.com/pointnetwork/cosmos-point-sdk/testutil/cli"
	"github.com/pointnetwork/cosmos-point-sdk/testutil/network"
	"github.com/pointnetwork/cosmos-point-sdk/x/authz/client/cli"
)

func CreateGrant(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := cli.NewCmdGrantAuthorization()
	clientCtx := val.ClientCtx
	return clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
}
