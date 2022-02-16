package testutil

import (
	"fmt"
	"strings"
	"time"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
)

func (s *IntegrationTestSuite) TestQueryAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := ExecGrant(
		val,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		expErrMsg string
	}{
		{
			"Error: Invalid grantee",
			[]string{
				val.Address.String(),
				"invalid grantee",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Error: Invalid granter",
			[]string{
				"invalid granter",
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Valid txn (json)",
			[]string{
				val.Address.String(),
				grantee.String(),
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			``,
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryGrants()
			clientCtx := val.ClientCtx
			resp, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(string(resp.Bytes()), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				var grants authz.QueryGrantsResponse
				err = val.ClientCtx.Codec.UnmarshalJSON(resp.Bytes(), &grants)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryAuthorization() {
	val := s.network.Validators[0]

	grantee := s.grantee
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := ExecGrant(
		val,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100steak", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"Error: Invalid grantee",
			[]string{
				val.Address.String(),
				"invalid grantee",
				typeMsgSend,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"",
		},
		{
			"Error: Invalid granter",
			[]string{
				"invalid granter",
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"",
		},
		{
			"no authorization found",
			[]string{
				val.Address.String(),
				grantee.String(),
				"typeMsgSend",
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			true,
			"",
		},
		{
			"Valid txn (json)",
			[]string{
				val.Address.String(),
				grantee.String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			},
			false,
			`{"@type":"/cosmos.bank.v1beta1.SendAuthorization","spend_limit":[{"denom":"steak","amount":"100"}]}`,
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryGrants()
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Contains(strings.TrimSpace(out.String()), tc.expectedOutput)
			}
		})
	}
}
