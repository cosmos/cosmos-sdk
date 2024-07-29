package simulation_test

import (
	"context"
	"math/rand"
	"testing"

	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/mint/simulation"
	"cosmossdk.io/x/mint/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestProposalMsgs(t *testing.T) {
	ac := codectestutil.CodecOptions{}.GetAddressCodec()
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	assert.Assert(t, len(weightedProposalMsgs) == 1)

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	assert.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	assert.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	msg, err := w0.MsgSimulatorFn()(context.Background(), r, accounts, ac)
	assert.NilError(t, err)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	authority, err := ac.BytesToString(address.Module("gov"))
	assert.NilError(t, err)

	assert.Equal(t, authority, msgUpdateParams.Authority)
	assert.Equal(t, uint64(122877), msgUpdateParams.Params.BlocksPerYear)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(95, 2), msgUpdateParams.Params.GoalBonded)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(94, 2), msgUpdateParams.Params.InflationMax)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(23, 2), msgUpdateParams.Params.InflationMin)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(89, 2), msgUpdateParams.Params.InflationRateChange)
	assert.Equal(t, "XhhuTSkuxK", msgUpdateParams.Params.MintDenom)
}
