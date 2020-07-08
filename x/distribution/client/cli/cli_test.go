package cli_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testutil.Config
	network *testutil.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	s.network = testutil.NewTestNetwork(s.T(), cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// import (
// 	"path/filepath"
// 	"testing"

// 	"github.com/stretchr/testify/require"
// 	tmtypes "github.com/tendermint/tendermint/types"

// 	"github.com/cosmos/cosmos-sdk/tests"
// 	"github.com/cosmos/cosmos-sdk/tests/cli"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/x/distribution/client/testutil"
// 	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
// )

// func TestCLIWithdrawRewards(t *testing.T) {
// 	t.SkipNow() // TODO: Bring back once viper is refactored.
// 	t.Parallel()
// 	f := cli.InitFixtures(t)

// 	genesisState := f.GenesisState()
// 	inflationMin := sdk.MustNewDecFromStr("1.0")
// 	var mintData minttypes.GenesisState
// 	f.Cdc.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData)
// 	mintData.Minter.Inflation = inflationMin
// 	mintData.Params.InflationMin = inflationMin
// 	mintData.Params.InflationMax = sdk.MustNewDecFromStr("1.0")
// 	mintDataBz, err := f.Cdc.MarshalJSON(mintData)
// 	require.NoError(t, err)
// 	genesisState[minttypes.ModuleName] = mintDataBz

// 	genFile := filepath.Join(f.SimdHome, "config", "genesis.json")
// 	genDoc, err := tmtypes.GenesisDocFromFile(genFile)
// 	require.NoError(t, err)
// 	genDoc.AppState, err = f.Cdc.MarshalJSON(genesisState)
// 	require.NoError(t, genDoc.SaveAs(genFile))

// 	// start simd server
// 	proc := f.SDStart()
// 	t.Cleanup(func() { proc.Stop(false) })

// 	params := testutil.QueryParameters(f)
// 	require.NotEmpty(t, params)

// 	fooAddr := f.KeyAddress(cli.KeyFoo)
// 	barAddr := f.KeyAddress(cli.KeyBar)
// 	fooVal := sdk.ValAddress(fooAddr)

// 	outstandingRewards := testutil.QueryValidatorOutstandingRewards(f, fooVal.String())
// 	require.NotEmpty(t, outstandingRewards)
// 	require.False(t, outstandingRewards.Rewards.IsZero())

// 	commission := testutil.QueryCommission(f, fooVal.String())
// 	require.NotEmpty(t, commission)
// 	require.False(t, commission.Commission.IsZero())

// 	rewards := testutil.QueryRewards(f, fooAddr)
// 	require.Len(t, rewards.Rewards, 1)
// 	require.NotEmpty(t, rewards.Total)

// 	// withdrawing rewards of a delegation for a single validator
// 	success := testutil.TxWithdrawRewards(f, fooVal, fooAddr.String(), "-y")
// 	require.True(t, success)

// 	rewards = testutil.QueryRewards(f, fooAddr)
// 	require.Len(t, rewards.Rewards, 1)
// 	require.Len(t, rewards.Total, 1)

// 	// Setting up a new withdraw address
// 	success, stdout, stderr := testutil.TxSetWithdrawAddress(f, fooAddr.String(), barAddr.String(), "--generate-only")
// 	require.True(t, success)
// 	require.Empty(t, stderr)

// 	msg := cli.UnmarshalStdTx(t, f.Cdc, stdout)
// 	require.NotZero(t, msg.Fee.Gas)
// 	require.Len(t, msg.Msgs, 1)
// 	require.Len(t, msg.GetSignatures(), 0)

// 	success, _, stderr = testutil.TxSetWithdrawAddress(f, cli.KeyFoo, barAddr.String(), "-y")
// 	require.True(t, success)
// 	require.Empty(t, stderr)
// 	tests.WaitForNextNBlocksTM(1, f.Port)

// 	// Withdraw all delegation rewards from all validators
// 	success, stdout, stderr = testutil.TxWithdrawAllRewards(f, fooAddr.String(), "--generate-only")
// 	require.True(t, success)
// 	require.Empty(t, stderr)

// 	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
// 	require.NotZero(t, msg.Fee.Gas)
// 	require.Len(t, msg.Msgs, 1)
// 	require.Len(t, msg.GetSignatures(), 0)

// 	success, _, stderr = testutil.TxWithdrawAllRewards(f, cli.KeyFoo, "-y")
// 	require.True(t, success)
// 	require.Empty(t, stderr)
// 	tests.WaitForNextNBlocksTM(1, f.Port)

// 	newTokens := sdk.NewCoin(cli.Denom, sdk.TokensFromConsensusPower(1))

// 	// Withdraw all delegation rewards from all validators
// 	success, stdout, stderr = testutil.TxFundCommunityPool(f, fooAddr.String(), newTokens, "--generate-only")
// 	require.True(t, success)
// 	require.Empty(t, stderr)

// 	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
// 	require.NotZero(t, msg.Fee.Gas)
// 	require.Len(t, msg.Msgs, 1)
// 	require.Len(t, msg.GetSignatures(), 0)

// 	success, _, stderr = testutil.TxFundCommunityPool(f, cli.KeyFoo, newTokens, "-y")
// 	require.True(t, success)
// 	require.Empty(t, stderr)
// 	tests.WaitForNextNBlocksTM(1, f.Port)

// 	amount := testutil.QueryCommunityPool(f)
// 	require.False(t, amount.IsZero())

// 	slashes := testutil.QuerySlashes(f, fooVal.String())
// 	require.Nil(t, slashes, nil)

// 	f.Cleanup()
// }
