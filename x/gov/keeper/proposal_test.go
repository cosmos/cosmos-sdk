package keeper

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
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

func (validProposal) GetTitle() string       { return "title" }
func (validProposal) GetDescription() string { return "description" }
func (validProposal) ProposalRoute() string  { return types.RouterKey }
func (validProposal) ProposalType() string   { return types.ProposalTypeText }
func (validProposal) String() string         { return "" }
func (validProposal) ValidateBasic() error   { return nil }

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

func (invalidProposalValidation) ValidateBasic() error {
	return errors.New("invalid proposal")
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
		expectedErr error
	}{
		{validProposal{}, nil},
		// Keeper does not check the validity of title and description, no error
		{invalidProposalTitle1{}, nil},
		{invalidProposalTitle2{}, nil},
		{invalidProposalDesc1{}, nil},
		{invalidProposalDesc2{}, nil},
		// error only when invalid route
		{invalidProposalRoute{}, types.ErrNoProposalHandlerExists},
		// Keeper does not call ValidateBasic, msg.ValidateBasic does
		{invalidProposalValidation{}, nil},
	}

	for i, tc := range testCases {
		_, err := keeper.SubmitProposal(ctx, tc.content)
		require.True(t, errors.Is(tc.expectedErr, err), "tc #%d; got: %v, expected: %v", i, err, tc.expectedErr)
	}
}

func TestGetProposalsFiltered(t *testing.T) {
	proposalID := uint64(1)
	ctx, _, keeper, _, _ := createTestInput(t, false, 100)
	status := []types.ProposalStatus{types.StatusDepositPeriod, types.StatusVotingPeriod}

	addr1 := sdk.AccAddress("foo")

	for _, s := range status {
		for i := 0; i < 50; i++ {
			p := types.NewProposal(TestProposal, proposalID, time.Now(), time.Now())
			p.Status = s

			if i%2 == 0 {
				d := types.NewDeposit(proposalID, addr1, nil)
				v := types.NewVote(proposalID, addr1, types.OptionYes)
				keeper.SetDeposit(ctx, d)
				keeper.SetVote(ctx, v)
			}

			keeper.SetProposal(ctx, p)
			proposalID++
		}
	}

	testCases := []struct {
		params             types.QueryProposalsParams
		expectedNumResults int
	}{
		{types.NewQueryProposalsParams(1, 50, types.StatusNil, nil, nil), 50},
		{types.NewQueryProposalsParams(1, 50, types.StatusDepositPeriod, nil, nil), 50},
		{types.NewQueryProposalsParams(1, 50, types.StatusVotingPeriod, nil, nil), 50},
		{types.NewQueryProposalsParams(1, 25, types.StatusNil, nil, nil), 25},
		{types.NewQueryProposalsParams(2, 25, types.StatusNil, nil, nil), 25},
		{types.NewQueryProposalsParams(1, 50, types.StatusRejected, nil, nil), 0},
		{types.NewQueryProposalsParams(1, 50, types.StatusNil, addr1, nil), 50},
		{types.NewQueryProposalsParams(1, 50, types.StatusNil, nil, addr1), 50},
		{types.NewQueryProposalsParams(1, 50, types.StatusNil, addr1, addr1), 50},
		{types.NewQueryProposalsParams(1, 50, types.StatusDepositPeriod, addr1, addr1), 25},
		{types.NewQueryProposalsParams(1, 50, types.StatusDepositPeriod, nil, nil), 50},
		{types.NewQueryProposalsParams(1, 50, types.StatusVotingPeriod, nil, nil), 50},
	}

	for _, tc := range testCases {
		proposals := keeper.GetProposalsFiltered(ctx, tc.params)
		require.Len(t, proposals, tc.expectedNumResults)

		for _, p := range proposals {
			if len(tc.params.ProposalStatus.String()) != 0 {
				require.Equal(t, tc.params.ProposalStatus, p.Status)
			}
		}
	}
}
