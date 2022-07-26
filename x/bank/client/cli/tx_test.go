package cli_test

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

func (s *CLITestSuite) TestSendTxCmd() {
	records := s.createKeyringRecords(1)

	addr1, err := records[0].GetAddress()
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		from, to  sdk.AccAddress
		amount    sdk.Coins
		extraArgs []string
		expectErr bool
	}{
		{
			"valid transaction",
			func() client.Context {
				return s.baseCtx
			},
			addr1,
			addr1,
			sdk.NewCoins(
				sdk.NewCoin("stake", sdk.NewInt(10)),
				sdk.NewCoin("photon", sdk.NewInt(40)),
			),
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", sdk.NewInt(10))).String()),
				fmt.Sprintf("--%s=test-chain", flags.FlagChainID),
			},
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			cmd := cli.NewSendTxCmd()
			ctx := svrcmd.CreateExecuteContext(context.Background())
			cmd.SetContext(ctx)
			s.Require().NoError(client.SetCmdClientContextHandler(tc.ctxGen(), cmd))

			cmd.SetArgs(append([]string{tc.from.String(), tc.to.String(), tc.amount.String()}, tc.extraArgs...))

			err := cmd.Execute()
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
