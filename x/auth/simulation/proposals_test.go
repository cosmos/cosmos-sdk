package simulation_test

import (
	"math/rand"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/x/auth/simulation"
	"cosmossdk.io/x/auth/types"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func TestProposalMsgs(t *testing.T) {
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

	msg, err := w0.MsgSimulatorFn()(r, accounts, codectestutil.CodecOptions{}.GetAddressCodec())
	assert.NilError(t, err)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	assert.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgUpdateParams.Authority)
	assert.Equal(t, uint64(999), msgUpdateParams.Params.MaxMemoCharacters)
	assert.Equal(t, uint64(905), msgUpdateParams.Params.TxSigLimit)
	assert.Equal(t, uint64(151), msgUpdateParams.Params.TxSizeCostPerByte)
	assert.Equal(t, uint64(213), msgUpdateParams.Params.SigVerifyCostED25519)
	assert.Equal(t, uint64(539), msgUpdateParams.Params.SigVerifyCostSecp256k1)
}
