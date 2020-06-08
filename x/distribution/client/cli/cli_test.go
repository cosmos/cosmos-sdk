// +build cli_test

package cli_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/mint"
)

func TestCLIWithdrawRewards(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	genesisState := f.GenesisState()
	inflationMin := sdk.MustNewDecFromStr("1.0")
	var mintData mint.GenesisState
	f.Cdc.UnmarshalJSON(genesisState[mint.ModuleName], &mintData)
	mintData.Minter.Inflation = inflationMin
	mintData.Params.InflationMin = inflationMin
	mintData.Params.InflationMax = sdk.MustNewDecFromStr("1.0")
	mintDataBz, err := f.Cdc.MarshalJSON(mintData)
	require.NoError(t, err)
	genesisState[mint.ModuleName] = mintDataBz

	genFile := filepath.Join(f.SimdHome, "config", "genesis.json")
	genDoc, err := tmtypes.GenesisDocFromFile(genFile)
	require.NoError(t, err)
	genDoc.AppState, err = f.Cdc.MarshalJSON(genesisState)
	require.NoError(t, genDoc.SaveAs(genFile))

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	params := testutil.QueryParameters(f)
	require.NotEmpty(t, params)

	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)
	fooVal := sdk.ValAddress(fooAddr)

	outStandingRewards := testutil.QueryValidatorOutstandingRewards(f, fooVal.String())
	require.NotEmpty(t, outStandingRewards)
	require.False(t, outStandingRewards.Rewards.IsZero())

	commission := testutil.QueryCommission(f, fooVal.String())
	require.NotEmpty(t, commission)
	require.False(t, commission.Commission.IsZero())

	rewards := testutil.QueryRewards(f, fooAddr)
	require.Equal(t, 1, len(rewards.Rewards))
	require.NotEmpty(t, rewards.Total)

	// withdrawing rewards of a delegation for a single validator
	success := testutil.TxWithdrawRewards(f, fooVal, fooAddr.String(), "-y")
	require.True(t, success)

	rewards = testutil.QueryRewards(f, fooAddr)
	require.Equal(t, 1, len(rewards.Rewards))
	require.Nil(t, rewards.Total)

	// Setting up a new withdraw address
	success, stdout, stderr := testutil.TxSetWithdrawAddress(f, fooAddr.String(), barAddr.String(), "--generate-only")
	require.True(f.T, success)
	require.Empty(f.T, stderr)

	msg := cli.UnmarshalStdTx(f.T, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

	success, _, stderr = testutil.TxSetWithdrawAddress(f, cli.KeyFoo, barAddr.String(), "-y")
	require.True(f.T, success)
	require.Empty(f.T, stderr)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Withdraw all delegation rewards from all validators
	success, stdout, stderr = testutil.TxWithdrawAllRewards(f, fooAddr.String(), "--generate-only")
	require.True(f.T, success)
	require.Empty(f.T, stderr)

	msg = cli.UnmarshalStdTx(f.T, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	success, _, stderr = testutil.TxWithdrawAllRewards(f, cli.KeyFoo, "-y")
	require.True(f.T, success)
	require.Empty(f.T, stderr)
	tests.WaitForNextNBlocksTM(1, f.Port)

	slashes := testutil.QuerySlashes(f, fooVal.String())
	require.Nil(t, slashes, nil)

	f.Cleanup()
}
