package cli_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
)

func (s *CLITestSuite) TestNewUnjailTxCmd() {
	val := s.addr
	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
	}{
		{
			"valid transaction",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync), // sync mode as there are no funds yet
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))).String()),
			},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.NewUnjailTxCmd()
			clientCtx := s.clientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErrMsg != "" {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				txResp := &sdk.TxResponse{}
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), txResp), out.String())
			}
		})
	}
}
