package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/staking/simulation"
	"cosmossdk.io/x/staking/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)
	addressCodec := codectestutil.CodecOptions{}.GetAddressCodec()
	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	assert.Assert(t, len(weightedProposalMsgs) == 1)

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	assert.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	assert.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	msg, err := w0.MsgSimulatorFn()(r, accounts, addressCodec)
	assert.NilError(t, err)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	addr, err := addressCodec.BytesToString(address.Module("gov"))
	assert.NilError(t, err)

	assert.Equal(t, addr, msgUpdateParams.Authority)
	assert.Equal(t, "GqiQWIXnku", msgUpdateParams.Params.BondDenom)
	assert.Equal(t, uint32(213), msgUpdateParams.Params.MaxEntries)
	assert.Equal(t, uint32(300), msgUpdateParams.Params.HistoricalEntries)
	assert.Equal(t, uint32(539), msgUpdateParams.Params.MaxValidators)
	assert.Equal(t, 8898194435*time.Second, msgUpdateParams.Params.UnbondingTime)
	assert.DeepEqual(t, sdkmath.LegacyNewDecWithPrec(579040435581502128, 18), msgUpdateParams.Params.MinCommissionRate)
}
