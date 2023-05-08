package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// SubmitProposal creates a new proposal given an array of messages
func (keeper Keeper) SubmitProposal(ctx context.Context, messages []sdk.Msg, metadata, title, summary string, proposer sdk.AccAddress, expedited bool) (v1.Proposal, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
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
		if m, ok := msg.(sdk.HasValidateBasic); ok {
			if err := m.ValidateBasic(); err != nil {
				return v1.Proposal{}, errorsmod.Wrap(types.ErrInvalidProposalMsg, err.Error())
			}
		}

		signers, _, err := keeper.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return v1.Proposal{}, err
		}
		if len(signers) != 1 {
			return v1.Proposal{}, types.ErrInvalidSigner
		}

		// assert that the governance module account is the only signer of the messages
		if !bytes.Equal(signers[0], keeper.GetGovernanceAccount(ctx).GetAddress()) {
			return v1.Proposal{}, errorsmod.Wrapf(types.ErrInvalidSigner, sdk.AccAddress(signers[0]).String())
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
			cacheCtx, _ := sdkCtx.CacheContext()
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

	params, err := keeper.GetParams(ctx)
	if err != nil {
		return v1.Proposal{}, err
	}

	submitTime := sdkCtx.BlockHeader().Time
	depositPeriod := params.MaxDepositPeriod

	proposal, err := v1.NewProposal(messages, proposalID, submitTime, submitTime.Add(*depositPeriod), metadata, title, summary, proposer, expedited)
	if err != nil {
		return v1.Proposal{}, err
	}

	keeper.SetProposal(ctx, proposal)
	keeper.InsertInactiveProposalQueue(ctx, proposalID, *proposal.DepositEndTime)
	keeper.SetProposalID(ctx, proposalID+1)

	// called right after a proposal is submitted
	keeper.Hooks().AfterProposalSubmission(ctx, proposalID)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute(types.AttributeKeyProposalMessages, msgsStr),
		),
	)

	return proposal, nil
}

// CancelProposal will cancel proposal before the voting period ends
func (keeper Keeper) CancelProposal(ctx context.Context, proposalID uint64, proposer string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	proposal, err := keeper.GetProposal(ctx, proposalID)
	if err != nil {
		return err
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
	if proposal.VotingEndTime != nil && proposal.VotingEndTime.Before(sdkCtx.BlockTime()) {
		return types.ErrVotingPeriodEnded.Wrapf("voting period is already ended for this proposal %d", proposalID)
	}

	// burn the (deposits * proposal_cancel_rate) amount or sent to cancellation destination address.
	// and deposits * (1 - proposal_cancel_rate) will be sent to depositors.
	params, err := keeper.GetParams(ctx)
	if err != nil {
		return err
	}

	err = keeper.ChargeDeposit(ctx, proposal.Id, params.ProposalCancelDest, params.ProposalCancelRatio)
	if err != nil {
		return err
	}

	if proposal.VotingStartTime != nil {
		err = keeper.deleteVotes(ctx, proposal.Id)
		if err != nil {
			return err
		}
	}

	err = keeper.DeleteProposal(ctx, proposal.Id)
	if err != nil {
		return err
	}

	keeper.Logger(ctx).Info(
		"proposal is canceled by proposer",
		"proposal", proposal.Id,
		"proposer", proposal.Proposer,
	)

	return nil
}

// GetProposal gets a proposal from store by ProposalID.
func (keeper Keeper) GetProposal(ctx context.Context, proposalID uint64) (proposal v1.Proposal, err error) {
	store := keeper.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.ProposalKey(proposalID))
	if err != nil {
		return
	}

	if bz == nil {
		return proposal, types.ErrProposalNotFound.Wrapf("proposal %d doesn't exist", proposalID)
	}

	if err = keeper.UnmarshalProposal(bz, &proposal); err != nil {
		return
	}

	return proposal, nil
}

// SetProposal sets a proposal to store.
func (keeper Keeper) SetProposal(ctx context.Context, proposal v1.Proposal) error {
	bz, err := keeper.MarshalProposal(proposal)
	if err != nil {
		return err
	}

	store := keeper.storeService.OpenKVStore(ctx)

	if proposal.Status == v1.StatusVotingPeriod {
		err = store.Set(types.VotingPeriodProposalKey(proposal.Id), []byte{1})
		if err != nil {
			return err
		}
	} else {
		err = store.Delete(types.VotingPeriodProposalKey(proposal.Id))
		if err != nil {
			return err
		}
	}

	return store.Set(types.ProposalKey(proposal.Id), bz)
}

// DeleteProposal deletes a proposal from store.
func (keeper Keeper) DeleteProposal(ctx context.Context, proposalID uint64) error {
	store := keeper.storeService.OpenKVStore(ctx)
	proposal, err := keeper.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	if proposal.DepositEndTime != nil {
		err := keeper.RemoveFromInactiveProposalQueue(ctx, proposalID, *proposal.DepositEndTime)
		if err != nil {
			return err
		}
	}
	if proposal.VotingEndTime != nil {
		err := keeper.RemoveFromActiveProposalQueue(ctx, proposalID, *proposal.VotingEndTime)
		if err != nil {
			return err
		}

		err = store.Delete(types.VotingPeriodProposalKey(proposalID))
		if err != nil {
			return err
		}
	}

	return store.Delete(types.ProposalKey(proposalID))
}

// IterateProposals iterates over all the proposals and performs a callback function.
func (keeper Keeper) IterateProposals(ctx context.Context, cb func(proposal v1.Proposal) error) error {
	store := keeper.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.ProposalsKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var proposal v1.Proposal
		err := keeper.UnmarshalProposal(iterator.Value(), &proposal)
		if err != nil {
			return err
		}

		err = cb(proposal)
		// exit early without error if cb returns ErrStopIterating
		if errorsmod.IsOf(err, errorsmod.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// GetProposals returns all the proposals from store
func (keeper Keeper) GetProposals(ctx context.Context) (proposals v1.Proposals, err error) {
	err = keeper.IterateProposals(ctx, func(proposal v1.Proposal) error {
		proposals = append(proposals, &proposal)
		return nil
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
func (keeper Keeper) GetProposalsFiltered(ctx context.Context, params v1.QueryProposalsParams) (v1.Proposals, error) {
	proposals, err := keeper.GetProposals(ctx)
	if err != nil {
		return nil, err
	}

	filteredProposals := make([]*v1.Proposal, 0, len(proposals))

	for _, p := range proposals {
		matchVoter, matchDepositor, matchStatus := true, true, true

		// match status (if supplied/valid)
		if v1.ValidProposalStatus(params.ProposalStatus) {
			matchStatus = p.Status == params.ProposalStatus
		}

		// match voter address (if supplied)
		if len(params.Voter) > 0 {
			_, err = keeper.GetVote(ctx, p.Id, params.Voter)
			// if no error, vote found, matchVoter = true
			matchVoter = err == nil
		}

		// match depositor (if supplied)
		if len(params.Depositor) > 0 {
			_, err = keeper.GetDeposit(ctx, p.Id, params.Depositor)
			// if no error, deposit found, matchDepositor = true
			matchDepositor = err == nil
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

	return filteredProposals, nil
}

// GetProposalID gets the highest proposal ID
func (keeper Keeper) GetProposalID(ctx context.Context) (proposalID uint64, err error) {
	store := keeper.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ProposalIDKey)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, errorsmod.Wrap(types.ErrInvalidGenesis, "initial proposal ID hasn't been set")
	}

	proposalID = types.GetProposalIDFromBytes(bz)
	return proposalID, nil
}

// SetProposalID sets the new proposal ID to the store
func (keeper Keeper) SetProposalID(ctx context.Context, proposalID uint64) error {
	store := keeper.storeService.OpenKVStore(ctx)
	return store.Set(types.ProposalIDKey, types.GetProposalIDBytes(proposalID))
}

// ActivateVotingPeriod activates the voting period of a proposal
func (keeper Keeper) ActivateVotingPeriod(ctx context.Context, proposal v1.Proposal) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	startTime := sdkCtx.BlockHeader().Time
	proposal.VotingStartTime = &startTime
	var votingPeriod *time.Duration
	params, err := keeper.GetParams(ctx)
	if err != nil {
		return err
	}

	if proposal.Expedited {
		votingPeriod = params.ExpeditedVotingPeriod
	} else {
		votingPeriod = params.VotingPeriod
	}
	endTime := proposal.VotingStartTime.Add(*votingPeriod)
	proposal.VotingEndTime = &endTime
	proposal.Status = v1.StatusVotingPeriod
	err = keeper.SetProposal(ctx, proposal)
	if err != nil {
		return err
	}

	err = keeper.RemoveFromInactiveProposalQueue(ctx, proposal.Id, *proposal.DepositEndTime)
	if err != nil {
		return err
	}

	return keeper.InsertActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)
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
