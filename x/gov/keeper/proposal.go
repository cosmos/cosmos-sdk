package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// SubmitProposal create new proposal given a content
func (keeper Keeper) SubmitProposal(ctx sdk.Context, content types.Content) (types.Proposal, error) {
	if !keeper.router.HasRoute(content.ProposalRoute()) {
		return nil, sdkerrors.Wrap(types.ErrNoProposalHandlerExists, content.ProposalRoute())
	}

	// Execute the proposal content in a cache-wrapped context to validate the
	// actual parameter changes before the proposal proceeds through the
	// governance process. State is not persisted.
	cacheCtx, _ := ctx.CacheContext()
	handler := keeper.router.GetRoute(content.ProposalRoute())
	if err := handler(cacheCtx, content); err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidProposalContent, err.Error())
	}

	proposalID, err := keeper.GetProposalID(ctx)
	if err != nil {
		return nil, err
	}

	submitTime := ctx.BlockHeader().Time
	depositPeriod := keeper.GetDepositParams(ctx).MaxDepositPeriod

	proposal := keeper.proposalCtr()
	err = proposal.SetContent(content)
	if err != nil {
		return nil, err
	}
	proposal.SetProposalID(proposalID)
	proposal.SetSubmitTime(submitTime)
	proposal.SetDepositEndTime(submitTime.Add(depositPeriod))

	keeper.SetProposal(ctx, proposal)
	keeper.InsertInactiveProposalQueue(ctx, proposalID, proposal.GetDepositEndTime())
	keeper.SetProposalID(ctx, proposalID+1)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return proposal, nil
}

// GetProposal get proposal from store by ProposalID
func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (proposal types.Proposal, ok bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.ProposalKey(proposalID))
	if bz == nil {
		return
	}
	proposal = keeper.proposalCtr()
	err := proposal.Unmarshal(bz)
	if err != nil {
		return nil, false
	}
	return proposal, true
}

// SetProposal set a proposal to store
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal types.Proposal) {
	store := ctx.KVStore(keeper.storeKey)
	proposal = keeper.proposalCtr()
	bz, err := proposal.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(types.ProposalKey(proposal.GetProposalID()), bz)
}

// DeleteProposal deletes a proposal from store
func (keeper Keeper) DeleteProposal(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		panic(fmt.Sprintf("couldn't find proposal with id#%d", proposalID))
	}
	keeper.RemoveFromInactiveProposalQueue(ctx, proposalID, proposal.GetDepositEndTime())
	keeper.RemoveFromActiveProposalQueue(ctx, proposalID, proposal.GetVotingEndTime())
	store.Delete(types.ProposalKey(proposalID))
}

// IterateProposals iterates over the all the proposals and performs a callback function
func (keeper Keeper) IterateProposals(ctx sdk.Context, cb func(proposal types.Proposal) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ProposalsKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposal := keeper.proposalCtr()
		err := proposal.Unmarshal(iterator.Value())
		if err != nil {
			panic(err)
		}

		if cb(proposal) {
			break
		}
	}
}

// GetProposals returns all the proposals from store
func (keeper Keeper) GetProposals(ctx sdk.Context) (proposals types.Proposals) {
	keeper.IterateProposals(ctx, func(proposal types.Proposal) bool {
		proposals = append(proposals, proposal)
		return false
	})
	return
}

// GetProposalsFiltered retrieves proposals filtered by a given set of params which
// include pagination parameters along with voter and depositor addresses and a
// proposal status. The voter address will filter proposals by whether or not
// that address has voted on proposals. The depositor address will filter proposals
// by whether or not that address has deposited to them. Finally, status will filter
// proposals by status.
//
// NOTE: If no filters are provided, all proposals will be returned in paginated
// form.
func (keeper Keeper) GetProposalsFiltered(ctx sdk.Context, params types.QueryProposalsParams) []types.Proposal {
	proposals := keeper.GetProposals(ctx)
	filteredProposals := make([]types.Proposal, 0, len(proposals))

	for _, p := range proposals {
		matchVoter, matchDepositor, matchStatus := true, true, true

		// match status (if supplied/valid)
		if types.ValidProposalStatus(params.ProposalStatus) {
			matchStatus = p.GetStatus() == params.ProposalStatus
		}

		// match voter address (if supplied)
		if len(params.Voter) > 0 {
			_, matchVoter = keeper.GetVote(ctx, p.GetProposalID(), params.Voter)
		}

		// match depositor (if supplied)
		if len(params.Depositor) > 0 {
			_, matchDepositor = keeper.GetDeposit(ctx, p.GetProposalID(), params.Depositor)
		}

		if matchVoter && matchDepositor && matchStatus {
			filteredProposals = append(filteredProposals, p)
		}
	}

	start, end := client.Paginate(len(filteredProposals), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		filteredProposals = []types.Proposal{}
	} else {
		filteredProposals = filteredProposals[start:end]
	}

	return filteredProposals
}

// GetProposalID gets the highest proposal ID
func (keeper Keeper) GetProposalID(ctx sdk.Context) (proposalID uint64, err error) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.ProposalIDKey)
	if bz == nil {
		return 0, sdkerrors.Wrap(types.ErrInvalidGenesis, "initial proposal ID hasn't been set")
	}

	proposalID = types.GetProposalIDFromBytes(bz)
	return proposalID, nil
}

// SetProposalID sets the new proposal ID to the store
func (keeper Keeper) SetProposalID(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.ProposalIDKey, types.GetProposalIDBytes(proposalID))
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal types.Proposal) {
	startTime := ctx.BlockHeader().Time
	proposal.SetVotingStartTime(startTime)
	votingPeriod := keeper.GetVotingParams(ctx).VotingPeriod
	proposal.SetVotingEndTime(startTime.Add(votingPeriod))
	proposal.SetStatus(types.StatusVotingPeriod)
	keeper.SetProposal(ctx, proposal)

	keeper.RemoveFromInactiveProposalQueue(ctx, proposal.GetProposalID(), proposal.GetDepositEndTime())
	keeper.InsertActiveProposalQueue(ctx, proposal.GetProposalID(), proposal.GetVotingEndTime())
}
