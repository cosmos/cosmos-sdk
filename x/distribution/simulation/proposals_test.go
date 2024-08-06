package simulation_test

import (
	"context"
	"math/rand"
	"testing"

	"gotest.tools/v3/assert"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/distribution/simulation"
	"cosmossdk.io/x/distribution/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)
	addressCodec := codectestutil.CodecOptions{}.GetAddressCodec()
	accounts := simtypes.RandomAccounts(r, 3)

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

	addr, err := addressCodec.BytesToString(sdk.AccAddress(address.Module("gov")))
	assert.NilError(t, err)
	assert.Equal(t, addr, msgUpdateParams.Authority)
	assert.DeepEqual(t, sdkmath.LegacyNewDec(0), msgUpdateParams.Params.CommunityTax)
	assert.Equal(t, true, msgUpdateParams.Params.WithdrawAddrEnabled)
}
