//go:build e2e
// +build e2e

package mint

import (
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/client/cli"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"gotest.tools/v3/assert"
)

type fixture struct {
	cfg     network.Config
	network *network.Network
}

func initFixture(t *testing.T) (*fixture, func()) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1

	genesisState := cfg.GenesisState
	var mintData minttypes.GenesisState
	assert.NilError(t, cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	inflation := sdk.MustNewDecFromStr("1.0")
	mintData.Minter.Inflation = inflation
	mintData.Params.InflationMin = inflation
	mintData.Params.InflationMax = inflation

	mintDataBz, err := cfg.Codec.MarshalJSON(&mintData)
	assert.NilError(t, err)
	genesisState[minttypes.ModuleName] = mintDataBz
	cfg.GenesisState = genesisState

	network, err := network.New(t, t.TempDir(), cfg)
	assert.NilError(t, err)
	assert.NilError(t, network.WaitForNextBlock())

	return &fixture{
		cfg:     cfg,
		network: network,
	}, func() { network.Cleanup() }
}

func TestGetCmdQueryParams(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			`{"mint_denom":"stake","inflation_rate_change":"0.130000000000000000","inflation_max":"1.000000000000000000","inflation_min":"1.000000000000000000","goal_bonded":"0.670000000000000000","blocks_per_year":"6311520"}`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", flags.FlagOutput)},
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

		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetCmdQueryParams()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			assert.NilError(t, err)
			assert.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func TestGetCmdQueryInflation(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]

	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			`1.000000000000000000`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", flags.FlagOutput)},
			`1.000000000000000000`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetCmdQueryInflation()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			assert.NilError(t, err)
			assert.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func TestGetCmdQueryAnnualProvisions(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			`500000000.000000000000000000`,
		},
		{
			"text output",
			[]string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=text", flags.FlagOutput)},
			`500000000.000000000000000000`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetCmdQueryAnnualProvisions()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			assert.NilError(t, err)
			assert.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}
