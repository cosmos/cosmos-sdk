package cli_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/client/cli"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	network *testnet.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testnet.DefaultConfig()
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
	s.network = testnet.New(s.T(), cfg)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestGetCmdQueryParams() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`{"mint_denom":"stake","inflation_rate_change":"0.130000000000000000","inflation_max":"1.000000000000000000","inflation_min":"1.000000000000000000","goal_bonded":"0.670000000000000000","blocks_per_year":"6311520"}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`blocks_per_year: "6311520"
goal_bonded: "0.670000000000000000"
inflation_max: "1.000000000000000000"
inflation_min: "1.000000000000000000"
inflation_rate_change: "0.130000000000000000"
mint_denom: stake`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryParams()
			_, out := testutil.ApplyMockIO(cmd)

			clientCtx := val.ClientCtx.WithOutput(out)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

			out.Reset()
			cmd.SetArgs(tc.args)

			s.Require().NoError(cmd.ExecuteContext(ctx))
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryInflation() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`"1.000000000000000000"`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`"1.000000000000000000"`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryInflation()
			_, out := testutil.ApplyMockIO(cmd)

			clientCtx := val.ClientCtx.WithOutput(out)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

			out.Reset()
			cmd.SetArgs(tc.args)

			s.Require().NoError(cmd.ExecuteContext(ctx))
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdQueryAnnualProvisions() {
	val := s.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			`"500000000.000000000000000000"`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			`"500000000.000000000000000000"`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryAnnualProvisions()
			_, out := testutil.ApplyMockIO(cmd)

			clientCtx := val.ClientCtx.WithOutput(out)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

			out.Reset()
			cmd.SetArgs(tc.args)

			s.Require().NoError(cmd.ExecuteContext(ctx))
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
