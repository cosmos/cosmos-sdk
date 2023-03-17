package keeper

import (
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// SubmitProposal creates a new proposal given an array of messages
func (keeper Keeper) SubmitProposal(ctx sdk.Context, messages []sdk.Msg, metadata, title, summary string, proposer sdk.AccAddress, expedited bool) (v1.Proposal, error) {
	err := keeper.assertMetadataLength(metadata)
	if err != nil {
		return v1.Proposal{}, err
	}

	// assert summary is no longer than predefined max length of metadata
	err = keeper.assertMetadataLength(summary)
	if err != nil {
		return v1.Proposal{}, err
	}

	// assert title is no longer than predefined max length of metadata
	err = keeper.assertMetadataLength(title)
	if err != nil {
		return v1.Proposal{}, err
	}

	// Will hold a comma-separated string of all Msg type URLs.
	msgsStr := ""

	// Loop through all messages and confirm that each has a handler and the gov module account
	// as the only signer
	for _, msg := range messages {
		msgsStr += fmt.Sprintf(",%s", sdk.MsgTypeURL(msg))

		// perform a basic validation of the message
		if err := msg.ValidateBasic(); err != nil {
			return v1.Proposal{}, errorsmod.Wrap(types.ErrInvalidProposalMsg, err.Error())
		}

		signers := msg.GetSigners()
		if len(signers) != 1 {
			return v1.Proposal{}, types.ErrInvalidSigner
		}

		// assert that the governance module account is the only signer of the messages
		if !signers[0].Equals(keeper.GetGovernanceAccount(ctx).GetAddress()) {
			return v1.Proposal{}, errorsmod.Wrapf(types.ErrInvalidSigner, signers[0].String())
		}

		// use the msg service router to see that there is a valid route for that message.
		handler := keeper.router.Handler(msg)
		if handler == nil {
			return v1.Proposal{}, errorsmod.Wrap(types.ErrUnroutableProposalMsg, sdk.MsgTypeURL(msg))
		}

		// Only if it's a MsgExecLegacyContent do we try to execute the
		// proposal in a cached context.
		// For other Msgs, we do not verify the proposal messages any further.
		// They may fail upon execution.
		// ref: https://github.com/cosmos/cosmos-sdk/pull/10868#discussion_r784872842
		if msg, ok := msg.(*v1.MsgExecLegacyContent); ok {
			cacheCtx, _ := ctx.CacheContext()
			if _, err := handler(cacheCtx, msg); err != nil {
				if errors.Is(types.ErrNoProposalHandlerExists, err) {
					return v1.Proposal{}, err
				}
				return v1.Proposal{}, errorsmod.Wrap(types.ErrInvalidProposalContent, err.Error())
			}
		}

	}

	proposalID, err := keeper.GetProposalID(ctx)
	if err != nil {
		return v1.Proposal{}, err
	}

	submitTime := ctx.BlockHeader().Time
	depositPeriod := keeper.GetParams(ctx).MaxDepositPeriod

	proposal, err := v1.NewProposal(messages, proposalID, submitTime, submitTime.Add(*depositPeriod), metadata, title, summary, proposer, expedited)
	if err != nil {
		return v1.Proposal{}, err
	}

	keeper.SetProposal(ctx, proposal)
	keeper.InsertInactiveProposalQueue(ctx, proposalID, *proposal.DepositEndTime)
	keeper.SetProposalID(ctx, proposalID+1)

	// called right after a proposal is submitted
	keeper.Hooks().AfterProposalSubmission(ctx, proposalID)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute(types.AttributeKeyProposalMessages, msgsStr),
		),
	)

	return proposal, nil
}

// CancelProposal will cancel proposal before the voting period ends
func (keeper Keeper) CancelProposal(ctx sdk.Context, proposalID uint64, proposer string) error {
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return errorsmod.Wrapf(types.ErrProposalNotFound, "proposal_id %d", proposalID)
	}

	// Checking proposal have proposer or not because old proposal doesn't have proposer field,
	// https://github.com/cosmos/cosmos-sdk/blob/v0.46.2/proto/cosmos/gov/v1/gov.proto#L43
	if proposal.Proposer == "" {
		return types.ErrInvalidProposal.Wrapf("proposal %d doesn't have proposer %s, so cannot be canceled", proposalID, proposer)
	}

	// Check creator of the proposal
	if proposal.Proposer != proposer {
		return types.ErrInvalidProposer.Wrapf("invalid proposer %s", proposer)
	}

	// Check if proposal is active or not
	if (proposal.Status != v1.StatusDepositPeriod) && (proposal.Status != v1.StatusVotingPeriod) {
		return types.ErrInvalidProposal.Wrap("proposal should be in the deposit or voting period")
	}

	// Check proposal voting period is ended.
	if proposal.VotingEndTime != nil && proposal.VotingEndTime.Before(ctx.BlockTime()) {
		return types.ErrVotingPeriodEnded.Wrapf("voting period is already ended for this proposal %d", proposalID)
	}

	// burn the (deposits * proposal_cancel_rate) amount or sent to cancellation destination address.
	// and deposits * (1 - proposal_cancel_rate) will be sent to depositors.
	params := keeper.GetParams(ctx)
	err := keeper.ChargeDeposit(ctx, proposal.Id, params.ProposalCancelDest, params.ProposalCancelRatio)
	if err != nil {
		return err
	}

	if proposal.VotingStartTime != nil {
		keeper.deleteVotes(ctx, proposal.Id)
	}

	keeper.DeleteProposal(ctx, proposal.Id)

	keeper.Logger(ctx).Info(
		"proposal is canceled by proposer",
		"proposal", proposal.Id,
		"proposer", proposal.Proposer,
	)

	return nil
}

// GetProposal gets a proposal from store by ProposalID.
// Panics if can't unmarshal the proposal.
func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID uint64) (v1.Proposal, bool) {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.ProposalKey(proposalID))
	if bz == nil {
		return v1.Proposal{}, false
	}

	var proposal v1.Proposal
	if err := keeper.UnmarshalProposal(bz, &proposal); err != nil {
		panic(err)
	}

	return proposal, true
}

// SetProposal sets a proposal to store.
// Panics if can't marshal the proposal.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal v1.Proposal) {
	bz, err := keeper.MarshalProposal(proposal)
	if err != nil {
		panic(err)
	}

	store := ctx.KVStore(keeper.storeKey)

	if proposal.Status == v1.StatusVotingPeriod {
		store.Set(types.VotingPeriodProposalKey(proposal.Id), []byte{1})
	} else {
		store.Delete(types.VotingPeriodProposalKey(proposal.Id))
	}

	store.Set(types.ProposalKey(proposal.Id), bz)
}

// DeleteProposal deletes a proposal from store.
// Panics if the proposal doesn't exist.
func (keeper Keeper) DeleteProposal(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		panic(fmt.Sprintf("couldn't find proposal with id#%d", proposalID))
	}

	if proposal.DepositEndTime != nil {
		keeper.RemoveFromInactiveProposalQueue(ctx, proposalID, *proposal.DepositEndTime)
	}
	if proposal.VotingEndTime != nil {
		keeper.RemoveFromActiveProposalQueue(ctx, proposalID, *proposal.VotingEndTime)
		store.Delete(types.VotingPeriodProposalKey(proposalID))
	}

	store.Delete(types.ProposalKey(proposalID))
}

// IterateProposals iterates over all the proposals and performs a callback function.
// Panics when the iterator encounters a proposal which can't be unmarshaled.
func (keeper Keeper) IterateProposals(ctx sdk.Context, cb func(proposal v1.Proposal) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.ProposalsKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal v1.Proposal
		err := keeper.UnmarshalProposal(iterator.Value(), &proposal)
		if err != nil {
			panic(err)
		}

		if cb(proposal) {
			break
		}
	}
}

// GetProposals returns all the proposals from store
func (keeper Keeper) GetProposals(ctx sdk.Context) (proposals v1.Proposals) {
	keeper.IterateProposals(ctx, func(proposal v1.Proposal) bool {
		proposals = append(proposals, &proposal)
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
func (keeper Keeper) GetProposalsFiltered(ctx sdk.Context, params v1.QueryProposalsParams) v1.Proposals {
	proposals := keeper.GetProposals(ctx)
	filteredProposals := make([]*v1.Proposal, 0, len(proposals))

	for _, p := range proposals {
		matchVoter, matchDepositor, matchStatus := true, true, true

		// match status (if supplied/valid)
		if v1.ValidProposalStatus(params.ProposalStatus) {
			matchStatus = p.Status == params.ProposalStatus
		}

		// match voter address (if supplied)
		if len(params.Voter) > 0 {
			_, matchVoter = keeper.GetVote(ctx, p.Id, params.Voter)
		}

		// match depositor (if supplied)
		if len(params.Depositor) > 0 {
			_, matchDepositor = keeper.GetDeposit(ctx, p.Id, params.Depositor)
		}

		if matchVoter && matchDepositor && matchStatus {
			filteredProposals = append(filteredProposals, p)
		}
	}

	start, end := client.Paginate(len(filteredProposals), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		filteredProposals = []*v1.Proposal{}
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
		return 0, errorsmod.Wrap(types.ErrInvalidGenesis, "initial proposal ID hasn't been set")
	}

	proposalID = types.GetProposalIDFromBytes(bz)
	return proposalID, nil
}

// SetProposalID sets the new proposal ID to the store
func (keeper Keeper) SetProposalID(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.ProposalIDKey, types.GetProposalIDBytes(proposalID))
}

// ActivateVotingPeriod activates the voting period of a proposal
func (keeper Keeper) ActivateVotingPeriod(ctx sdk.Context, proposal v1.Proposal) {
	startTime := ctx.BlockHeader().Time
	proposal.VotingStartTime = &startTime
	var votingPeriod *time.Duration
	if proposal.Expedited {
		votingPeriod = keeper.GetParams(ctx).ExpeditedVotingPeriod
	} else {
		votingPeriod = keeper.GetParams(ctx).VotingPeriod
	}
	endTime := proposal.VotingStartTime.Add(*votingPeriod)
	proposal.VotingEndTime = &endTime
	proposal.Status = v1.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	keeper.RemoveFromInactiveProposalQueue(ctx, proposal.Id, *proposal.DepositEndTime)
	keeper.InsertActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)
}

// MarshalProposal marshals the proposal and returns binary encoded bytes.
func (keeper Keeper) MarshalProposal(proposal v1.Proposal) ([]byte, error) {
	bz, err := keeper.cdc.Marshal(&proposal)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

// UnmarshalProposal unmarshals the proposal.
func (keeper Keeper) UnmarshalProposal(bz []byte, proposal *v1.Proposal) error {
	err := keeper.cdc.Unmarshal(bz, proposal)
	if err != nil {
		return err
	}
	return nil
}
