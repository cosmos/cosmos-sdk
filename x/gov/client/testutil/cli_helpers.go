package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func SubmitTestProposal(clientCtx client.Context, addr string, bondDenom string) (testutil.BufferWriter, error) {
	args := []string{
		fmt.Sprintf("--%s='Text Proposal'", cli.FlagTitle),
		fmt.Sprintf("--%s='Where is the title!?'", cli.FlagDescription),
		fmt.Sprintf("--%s=%s", cli.FlagProposalType, types.ProposalTypeText),
		fmt.Sprintf("--%s=%s", cli.FlagDeposit, sdk.NewCoin(bondDenom, sdk.NewInt(5431)).String()),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, addr),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(10))).String()),
	}

	return clitestutil.ExecTestCLICmd(clientCtx, cli.NewCmdSubmitProposal(), args)
}
