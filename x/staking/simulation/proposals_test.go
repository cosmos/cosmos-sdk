package simulation_test

import (
	"context"
	"math/rand"
	"testing"

	"gotest.tools/v3/assert"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	accounts := simtypes.RandomAccounts(r, 3)
	addressCodec := codectestutil.CodecOptions{}.NewInterfaceRegistry().SigningContext().AddressCodec()
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

	assert.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgUpdateParams.Authority)
	assert.Equal(t, uint32(905), msgUpdateParams.Params.MaxEntries)
	assert.Equal(t, uint32(540), msgUpdateParams.Params.HistoricalEntries)
	assert.Equal(t, uint32(151), msgUpdateParams.Params.MaxValidators)
	assert.Equal(t, "2417694h42m25s", msgUpdateParams.Params.UnbondingTime.String())
}
