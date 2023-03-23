package cli_test

import (
	"context"
	"fmt"
	"io"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	cli "cosmossdk.io/x/circuit/client/cli"
)

func (s *CLITestSuite) TestAuthorizeCircuitBreakerCmd() {
	cmd := cli.AuthorizeCircuitBreakerCmd()
	cmd.SetOutput(io.Discard)

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		args      []string
		expectErr bool
	}{
		{
			name: "Authorize an account to trip the circuit breaker.",
			ctxGen: func() client.Context {
				return s.baseCtx
			},
			args: []string{
				s.accounts[1].String(),
				"2",
				"cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.Msg",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accounts[0].String()),
			},
			expectErr: false,
		},
		{
			name: "Authorize an account to trip the circuit breaker.",
			ctxGen: func() client.Context {
				return s.baseCtx
			},
			args: []string{
				s.accounts[2].String(),
				"3",
				"cosmos.bank.v1beta1.MsgSend",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.accounts[0].String()),
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			s.Require().NoError(client.SetCmdClientContextHandler(tc.ctxGen(), cmd))

			err := cmd.Execute()

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// func (s *CLITestSuite) TestTripCircuitBreakerCmd(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		cmd         *cobra.Command
// 		args        []string
// 		expectedErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tc := range tests {
// 		tc := tc
// 		var resp sdk.TxResponse
// 		s.Run(tc.name, func() {
// 			cmd := tc.cmd
// 			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)
// 			if tc.expectedErr {
// 				s.Require().Error(err)
// 			} else {
// 				s.Require().NoError(err)
// 				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
// 			}
// 		})
// 	}
// }
