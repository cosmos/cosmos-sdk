package testutil

import (
	"fmt"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/nft/client/cli"
)

func ExecSend(val *network.Validator, args []string) (testutil.BufferWriter, error) {
	cmd := cli.NewCmdSend()
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryClass(val *network.Validator, classID string) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQueryClass()
	var args []string
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryClasses(val *network.Validator) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQueryClasses()
	var args []string
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryNFT(val *network.Validator, classID, nftID string) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQueryNFT()
	var args []string
	args = append(args, classID)
	args = append(args, nftID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryNFTs(val *network.Validator, classID string) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQueryNFTs()
	var args []string
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryNFTsByOwner(val *network.Validator, classID, owner string) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQueryNFTs()
	var args []string
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=%s", cli.FlagOwner, owner))
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryOwner(val *network.Validator, classID, nftID string) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQueryOwner()
	var args []string
	args = append(args, classID)
	args = append(args, nftID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQueryBalance(val *network.Validator, classID, owner string) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQueryBalance()
	var args []string
	args = append(args, owner)
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}

func ExecQuerySupply(val *network.Validator, classID string) (testutil.BufferWriter, error) {
	cmd := cli.GetCmdQuerySupply()
	var args []string
	args = append(args, classID)
	args = append(args, fmt.Sprintf("--%s=json", tmcli.OutputFlag))
	return clitestutil.ExecTestCLICmd(val.ClientCtx, cmd, args)
}
