package authz

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	authzclitestutil "github.com/cosmos/cosmos-sdk/x/authz/client/testutil"
)

func (s *E2ETestSuite) TestQueryAuthorizations() {
	val := s.network.Validators[0]

	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

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
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Error: Invalid granter",
			[]string{
				"invalid granter",
				grantee.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"decoding bech32 failed: invalid character in string: ' '",
		},
		{
			"Valid txn (json)",
			[]string{
				val.Address.String(),
				grantee.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			``,
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryGrants(address.NewBech32Codec("cosmos"))
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

func (s *E2ETestSuite) TestQueryAuthorization() {
	val := s.network.Validators[0]

	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := authzclitestutil.CreateGrant(
		val.ClientCtx,
		[]string{
			grantee.String(),
			"send",
			fmt.Sprintf("--%s=100stake", cli.FlagSpendLimit),
			fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
			fmt.Sprintf("--%s=%d", cli.FlagExpiration, twoHours),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		},
	)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

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
				fmt.Sprintf("--%s=json", flags.FlagOutput),
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
				fmt.Sprintf("--%s=json", flags.FlagOutput),
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
				fmt.Sprintf("--%s=json", flags.FlagOutput),
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
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			`{"@type":"/cosmos.bank.v1beta1.SendAuthorization","spend_limit":[{"denom":"stake","amount":"100"}],"allow_list":[]}`,
		},
		{
			"Valid txn with allowed list (json)",
			[]string{
				val.Address.String(),
				s.grantee[3].String(),
				typeMsgSend,
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			fmt.Sprintf(`{"@type":"/cosmos.bank.v1beta1.SendAuthorization","spend_limit":[{"denom":"stake","amount":"88"}],"allow_list":["%s"]}`, s.grantee[4]),
		},
	}
	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryGrants(address.NewBech32Codec("cosmos"))
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

func (s *E2ETestSuite) TestQueryGranterGrants() {
	val := s.network.Validators[0]
	grantee := s.grantee[0]
	require := s.Require()

	testCases := []struct {
		name        string
		args        []string
		expectErr   bool
		expectedErr string
		expItems    int
	}{
		{
			"invalid address",
			[]string{
				"invalid-address",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			true,
			"decoding bech32 failed",
			0,
		},
		{
			"no authorization found",
			[]string{
				grantee.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"",
			0,
		},
		{
			"valid case",
			[]string{
				val.Address.String(),
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"",
			8,
		},
		{
			"valid case with pagination",
			[]string{
				val.Address.String(),
				"--limit=2",
				fmt.Sprintf("--%s=json", flags.FlagOutput),
			},
			false,
			"",
			2,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryGranterGrants(address.NewBech32Codec("cosmos"))
			clientCtx := val.ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				require.Error(err)
				require.Contains(out.String(), tc.expectedErr)
			} else {
				require.NoError(err)
				var grants authz.QueryGranterGrantsResponse
				require.NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &grants))
				require.Len(grants.Grants, tc.expItems)
			}
		})
	}
}
