package keeper_test

import (
	"testing"

	proto "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappcodec "github.com/cosmos/cosmos-sdk/simapp/codec"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestIncrementProposalNumber(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	tp := TestProposal
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	app.GovKeeper.SubmitProposal(ctx, tp)
	proposal6, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.Equal(t, uint64(6), proposal6.ProposalID)
}

func TestProposalQueues(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	appCodec := simappcodec.NewAppCodec(app.Codec())

	// create test proposals
	tp := TestProposal
	proposal, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	inactiveIterator := app.GovKeeper.InactiveProposalQueueIterator(ctx, proposal.DepositEndTime)
	require.True(t, inactiveIterator.Valid())

	proposalID := types.GetProposalIDFromBytes(inactiveIterator.Value())
	require.Equal(t, proposalID, proposal.ProposalID)
	inactiveIterator.Close()

	app.GovKeeper.ActivateVotingPeriod(ctx, proposal)

	proposal, ok := app.GovKeeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := app.GovKeeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())

	var propIDValue proto.UInt64Value
	appCodec.UnmarshalBinaryLengthPrefixed(activeIterator.Value(), &propIDValue)

	require.Equal(t, propIDValue.Value, proposal.ProposalID)
	activeIterator.Close()
}
