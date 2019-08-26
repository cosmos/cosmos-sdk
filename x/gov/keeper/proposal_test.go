package keeper

import (
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
)

func TestGetSetProposal(t *testing.T) {
	ctx, _, keeper, _, _ := createTestInput(t, false, 100)

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	keeper.SetProposal(ctx, proposal)

	gotProposal, ok := keeper.GetProposal(ctx, proposalID)
	require.True(t, ok)
	require.True(t, ProposalEqual(proposal, gotProposal))
}

func TestActivateVotingPeriod(t *testing.T) {
	ctx, _, keeper, _, _ := createTestInput(t, false, 100)

	tp := TestProposal
	proposal, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)

	require.True(t, proposal.VotingStartTime.Equal(time.Time{}))

	keeper.activateVotingPeriod(ctx, proposal)

	require.True(t, proposal.VotingStartTime.Equal(ctx.BlockHeader().Time))

	proposal, ok := keeper.GetProposal(ctx, proposal.ProposalID)
	require.True(t, ok)

	activeIterator := keeper.ActiveProposalQueueIterator(ctx, proposal.VotingEndTime)
	require.True(t, activeIterator.Valid())

	proposalID := types.GetProposalIDFromBytes(activeIterator.Value())
	require.Equal(t, proposalID, proposal.ProposalID)
	activeIterator.Close()
}

type validProposal struct{}

func (validProposal) GetTitle() string         { return "title" }
func (validProposal) GetDescription() string   { return "description" }
func (validProposal) ProposalRoute() string    { return types.RouterKey }
func (validProposal) ProposalType() string     { return types.ProposalTypeText }
func (validProposal) String() string           { return "" }
func (validProposal) ValidateBasic() sdk.Error { return nil }

type invalidProposalTitle1 struct{ validProposal }

func (invalidProposalTitle1) GetTitle() string { return "" }

type invalidProposalTitle2 struct{ validProposal }

func (invalidProposalTitle2) GetTitle() string { return strings.Repeat("1234567890", 100) }

type invalidProposalDesc1 struct{ validProposal }

func (invalidProposalDesc1) GetDescription() string { return "" }

type invalidProposalDesc2 struct{ validProposal }

func (invalidProposalDesc2) GetDescription() string { return strings.Repeat("1234567890", 1000) }

type invalidProposalRoute struct{ validProposal }

func (invalidProposalRoute) ProposalRoute() string { return "nonexistingroute" }

type invalidProposalValidation struct{ validProposal }

func (invalidProposalValidation) ValidateBasic() sdk.Error {
	return sdk.NewError(sdk.CodespaceUndefined, sdk.CodeInternal, "")
}

func registerTestCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(validProposal{}, "test/validproposal", nil)
	cdc.RegisterConcrete(invalidProposalTitle1{}, "test/invalidproposalt1", nil)
	cdc.RegisterConcrete(invalidProposalTitle2{}, "test/invalidproposalt2", nil)
	cdc.RegisterConcrete(invalidProposalDesc1{}, "test/invalidproposald1", nil)
	cdc.RegisterConcrete(invalidProposalDesc2{}, "test/invalidproposald2", nil)
	cdc.RegisterConcrete(invalidProposalRoute{}, "test/invalidproposalr", nil)
	cdc.RegisterConcrete(invalidProposalValidation{}, "test/invalidproposalv", nil)
}

func TestSubmitProposal(t *testing.T) {
	ctx, _, keeper, _, _ := createTestInput(t, false, 100)

	registerTestCodec(keeper.cdc)

	testCases := []struct {
		content     types.Content
		expectedErr sdk.Error
	}{
		{validProposal{}, nil},
		// Keeper does not check the validity of title and description, no error
		{invalidProposalTitle1{}, nil},
		{invalidProposalTitle2{}, nil},
		{invalidProposalDesc1{}, nil},
		{invalidProposalDesc2{}, nil},
		// error only when invalid route
		{invalidProposalRoute{}, types.ErrNoProposalHandlerExists(types.DefaultCodespace, invalidProposalRoute{})},
		// Keeper does not call ValidateBasic, msg.ValidateBasic does
		{invalidProposalValidation{}, nil},
	}

	for _, tc := range testCases {
		_, err := keeper.SubmitProposal(ctx, tc.content)
		require.Equal(t, tc.expectedErr, err, "unexpected type of error: %s", err)
	}
}
