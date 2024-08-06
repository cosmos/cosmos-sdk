package simulation_test

import (
	"context"
	"math/rand"
	"testing"

	"gotest.tools/v3/assert"

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

	msg, err := w0.MsgSimulatorFn()(context.Background(), r, accounts, addressCodec)
	assert.NilError(t, err)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	addr, err := addressCodec.BytesToString(address.Module("gov"))
	assert.NilError(t, err)

	assert.Equal(t, addr, msgUpdateParams.Authority)
	assert.Equal(t, "stake", msgUpdateParams.Params.BondDenom)
	assert.Equal(t, uint32(905), msgUpdateParams.Params.MaxEntries)
	assert.Equal(t, uint32(540), msgUpdateParams.Params.HistoricalEntries)
	assert.Equal(t, uint32(151), msgUpdateParams.Params.MaxValidators)
	assert.Equal(t, "2417694h42m25s", msgUpdateParams.Params.UnbondingTime.String())
}
