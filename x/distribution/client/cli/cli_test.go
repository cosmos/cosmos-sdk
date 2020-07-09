package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutil.Config
	network *testutil.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultConfig()
	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	var mintData minttypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	inflation := sdk.MustNewDecFromStr("1.0")
	mintData.Minter.Inflation = inflation
	mintData.Params.InflationMin = inflation
	mintData.Params.InflationMax = inflation

	mintDataBz, err := cfg.Codec.MarshalJSON(mintData)
	s.Require().NoError(err)
	genesisState[minttypes.ModuleName] = mintDataBz
	cfg.GenesisState = genesisState

	s.cfg = cfg
	s.network = testutil.NewTestNetwork(s.T(), cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetCmdQueryParams() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"default output",
			[]string{},
			`{"community_tax":"0.020000000000000000","base_proposer_reward":"0.010000000000000000","bonus_proposer_reward":"0.040000000000000000","withdraw_addr_enabled":true}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`base_proposer_reward: "0.010000000000000000"
bonus_proposer_reward: "0.040000000000000000"
community_tax: "0.020000000000000000"
withdraw_addr_enabled: true`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()

			cmd := cli.GetCmdQueryParams()
			cmd.SetErr(buf)
			cmd.SetOut(buf)
			cmd.SetArgs(tc.args)

			s.Require().NoError(cmd.ExecuteContext(ctx))
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(buf.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorOutstandingRewards() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	s.network.WaitForHeight(3)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				"foo",
			},
			true,
			"",
		},
		{
			"default output",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
			},
			false,
			`{"rewards":[{"denom":"stake","amount":"232.260000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
			},
			false,
			`rewards:
- amount: "232.260000000000000000"
  denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()

			cmd := cli.GetCmdQueryValidatorOutstandingRewards()
			cmd.SetErr(buf)
			cmd.SetOut(buf)
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(buf.String()))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorCommission() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	s.network.WaitForHeight(3)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				"foo",
			},
			true,
			"",
		},
		{
			"default output",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
			},
			false,
			`{"commission":[{"denom":"stake","amount":"232.260000000000000000"}]}`,
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(),
			},
			false,
			`commission:
- amount: "232.260000000000000000"
  denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()

			cmd := cli.GetCmdQueryValidatorCommission()
			cmd.SetErr(buf)
			cmd.SetOut(buf)
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(buf.String()))
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryValidatorSlashes() {
	buf := new(bytes.Buffer)
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx.WithOutput(buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	s.network.WaitForHeight(3)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				"foo", "1", "3",
			},
			true,
			"",
		},
		{
			"invalid start height",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "-1", "3",
			},
			true,
			"",
		},
		{
			"invalid end height",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "1", "-3",
			},
			true,
			"",
		},
		{
			"default output",
			[]string{
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "1", "3",
			},
			false,
			"null",
		},
		{
			"text output",
			[]string{
				fmt.Sprintf("--%s=text", tmcli.OutputFlag),
				fmt.Sprintf("--%s=3", flags.FlagHeight),
				sdk.ValAddress(val.Address).String(), "1", "3",
			},
			false,
			"null",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			buf.Reset()

			cmd := cli.GetCmdQueryValidatorSlashes()
			cmd.SetErr(buf)
			cmd.SetOut(buf)
			cmd.SetArgs(tc.args)

			err := cmd.ExecuteContext(ctx)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(buf.String()))
			}
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
