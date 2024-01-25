package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SubmitProposal creates a new proposal given an array of messages
func (k Keeper) SubmitProposal(ctx context.Context, messages []sdk.Msg, metadata, title, summary string, proposer sdk.AccAddress, proposalType v1.ProposalType) (v1.Proposal, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return v1.Proposal{}, err
	}

	// additional checks per proposal types
	switch proposalType {
	case v1.ProposalType_PROPOSAL_TYPE_OPTIMISTIC:
		proposerStr, _ := k.authKeeper.AddressCodec().BytesToString(proposer)
		if len(params.OptimisticAuthorizedAddresses) > 0 && !slices.Contains(params.OptimisticAuthorizedAddresses, proposerStr) {
			return v1.Proposal{}, errorsmod.Wrap(types.ErrInvalidProposer, "proposer is not authorized to submit optimistic proposal")
		}
	case v1.ProposalType_PROPOSAL_TYPE_MULTIPLE_CHOICE:
		if len(messages) > 0 { // cannot happen, except when the proposal is created via keeper call instead of message server.
			return v1.Proposal{}, errorsmod.Wrap(types.ErrInvalidProposalMsg, "multiple choice proposal should not contain any messages")
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msgs := []string{} // will hold a string slice of all Msg type URLs.

	// Loop through all messages and confirm that each has a handler and the gov module account as the only signer
	for _, msg := range messages {
		msgs = append(msgs, sdk.MsgTypeURL(msg))

		// perform a basic validation of the message
		if m, ok := msg.(sdk.HasValidateBasic); ok {
			if err := m.ValidateBasic(); err != nil {
				return v1.Proposal{}, errorsmod.Wrap(types.ErrInvalidProposalMsg, err.Error())
			}
		}

		signers, _, err := k.cdc.GetMsgV1Signers(msg)
		if err != nil {
			return v1.Proposal{}, err
		}
		if len(signers) != 1 {
			return v1.Proposal{}, types.ErrInvalidSigner
		}

		// assert that the governance module account is the only signer of the messages
		if !bytes.Equal(signers[0], k.GetGovernanceAccount(ctx).GetAddress()) {
			return v1.Proposal{}, errorsmod.Wrapf(types.ErrInvalidSigner, sdk.AccAddress(signers[0]).String())
		}

		// use the msg service router to see that there is a valid route for that message.
		handler := k.router.Handler(msg)
		if handler == nil {
			return v1.Proposal{}, errorsmod.Wrap(types.ErrUnroutableProposalMsg, sdk.MsgTypeURL(msg))
		}

		// Only if it's a MsgExecLegacyContent do we try to execute the
		// proposal in a cached context.
		// For other Msgs, we do not verify the proposal messages any further.
		// They may fail upon execution.
		// ref: https://github.com/cosmos/cosmos-sdk/pull/10868#discussion_r784872842
		msg, ok := msg.(*v1.MsgExecLegacyContent)
		if !ok {
			continue
		}
		cacheCtx, _ := sdkCtx.CacheContext()
		if _, err := handler(cacheCtx, msg); err != nil {
			if errors.Is(types.ErrNoProposalHandlerExists, err) {
				return v1.Proposal{}, err
			}
			return v1.Proposal{}, errorsmod.Wrap(types.ErrInvalidProposalContent, err.Error())
		}

	}

	proposalID, err := k.ProposalID.Next(ctx)
	if err != nil {
		return v1.Proposal{}, err
	}

	submitTime := sdkCtx.HeaderInfo().Time
	proposal, err := v1.NewProposal(messages, proposalID, submitTime, submitTime.Add(*params.MaxDepositPeriod), metadata, title, summary, proposer, proposalType)
	if err != nil {
		return v1.Proposal{}, err
	}

	err = k.SetProposal(ctx, proposal)
	if err != nil {
		return v1.Proposal{}, err
	}
	err = k.InactiveProposalsQueue.Set(ctx, collections.Join(*proposal.DepositEndTime, proposalID), proposalID)
	if err != nil {
		return v1.Proposal{}, err
	}

	// called right after a proposal is submitted
	err = k.Hooks().AfterProposalSubmission(ctx, proposalID)
	if err != nil {
		return v1.Proposal{}, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
			sdk.NewAttribute(types.AttributeKeyProposalMessages, strings.Join(msgs, ",")),
		),
	)

	return proposal, nil
}

// CancelProposal will cancel proposal before the voting period ends
func (k Keeper) CancelProposal(ctx context.Context, proposalID uint64, proposer string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	proposal, err := k.Proposals.Get(ctx, proposalID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrInvalidProposal.Wrapf("proposal %d doesn't exist", proposalID)
		}
		return err
	}

	params, err := k.Params.Get(ctx)
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

	// Check proposal is not too far in voting period to be canceled
	if proposal.VotingEndTime != nil {
		currentTime := sdkCtx.HeaderInfo().Time

		maxCancelPeriodRate := sdkmath.LegacyMustNewDecFromStr(params.ProposalCancelMaxPeriod)
		maxCancelPeriod := time.Duration(float64(proposal.VotingEndTime.Sub(*proposal.VotingStartTime)) * maxCancelPeriodRate.MustFloat64()).Round(time.Second)

		if proposal.VotingEndTime.Before(currentTime) {
			return types.ErrVotingPeriodEnded.Wrapf("voting period is already ended for this proposal %d", proposalID)
		} else if proposal.VotingEndTime.Add(-maxCancelPeriod).Before(currentTime) {
			return types.ErrTooLateToCancel.Wrapf("proposal %d is too late to cancel", proposalID)
		}
	}

	// burn the (deposits * proposal_cancel_rate) amount or sent to cancellation destination address.
	// and deposits * (1 - proposal_cancel_rate) will be sent to depositors.
	err = k.ChargeDeposit(ctx, proposal.Id, params.ProposalCancelDest, params.ProposalCancelRatio)
	if err != nil {
		return err
	}

	if proposal.VotingStartTime != nil {
		err = k.deleteVotes(ctx, proposal.Id)
		if err != nil {
			return err
		}
	}

	err = k.DeleteProposal(ctx, proposal.Id)
	if err != nil {
		return err
	}

	k.Logger(ctx).Info(
		"proposal is canceled by proposer",
		"proposal", proposal.Id,
		"proposer", proposal.Proposer,
	)

	return nil
}

// SetProposal sets a proposal to store.
func (k Keeper) SetProposal(ctx context.Context, proposal v1.Proposal) error {
	if proposal.Status == v1.StatusVotingPeriod {
		err := k.VotingPeriodProposals.Set(ctx, proposal.Id, []byte{1})
		if err != nil {
			return err
		}
	} else {
		err := k.VotingPeriodProposals.Remove(ctx, proposal.Id)
		if err != nil {
			return err
		}
	}

	return k.Proposals.Set(ctx, proposal.Id, proposal)
}

// DeleteProposal deletes a proposal from store.
func (k Keeper) DeleteProposal(ctx context.Context, proposalID uint64) error {
	proposal, err := k.Proposals.Get(ctx, proposalID)
	if err != nil {
		return err
	}

	if proposal.DepositEndTime != nil {
		err := k.InactiveProposalsQueue.Remove(ctx, collections.Join(*proposal.DepositEndTime, proposalID))
		if err != nil {
			return err
		}
	}
	if proposal.VotingEndTime != nil {
		err := k.ActiveProposalsQueue.Remove(ctx, collections.Join(*proposal.VotingEndTime, proposalID))
		if err != nil {
			return err
		}

		err = k.VotingPeriodProposals.Remove(ctx, proposalID)
		if err != nil {
			return err
		}
	}

	return k.Proposals.Remove(ctx, proposalID)
}

// ActivateVotingPeriod activates the voting period of a proposal
func (k Keeper) ActivateVotingPeriod(ctx context.Context, proposal v1.Proposal) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	startTime := sdkCtx.HeaderInfo().Time
	proposal.VotingStartTime = &startTime
	var votingPeriod *time.Duration
	params, err := k.Params.Get(ctx)
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
	err = k.SetProposal(ctx, proposal)
	if err != nil {
		return err
	}

	err = k.InactiveProposalsQueue.Remove(ctx, collections.Join(*proposal.DepositEndTime, proposal.Id))
	if err != nil {
		return err
	}

	return k.ActiveProposalsQueue.Set(ctx, collections.Join(*proposal.VotingEndTime, proposal.Id), proposal.Id)
}
