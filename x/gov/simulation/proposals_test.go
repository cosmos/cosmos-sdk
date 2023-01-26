package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	ctx := sdk.NewContext(nil, tmproto.Header{}, true, nil)
	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	require.Len(t, weightedProposalMsgs, 1)

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	require.Equal(t, simulation.OpWeightSubmitTextProposal, w0.AppParamsKey())
	require.Equal(t, simulation.DefaultWeightTextProposal, w0.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	require.Nil(t, msg)
}
