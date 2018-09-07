package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// Handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDeposit:
			return handleMsgDeposit(ctx, keeper, msg)
		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)
		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)
		default:
			errMsg := "Unrecognized gov msg type"
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {

	proposal := keeper.NewTextProposal(ctx, msg.Title, msg.Description, msg.ProposalType)

	err, votingStarted := keeper.AddDeposit(ctx, proposal.GetProposalID(), msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(proposal.GetProposalID())

	resTags := sdk.NewTags(
		tags.Action, tags.ActionSubmitProposal,
		tags.Proposer, []byte(msg.Proposer.String()),
		tags.ProposalID, proposalIDBytes,
	)

	if votingStarted {
		resTags.AppendTag(tags.VotingPeriodStart, proposalIDBytes)
	}

	return sdk.Result{
		Data: proposalIDBytes,
		Tags: resTags,
	}
}

func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {

	err, votingStarted := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(msg.ProposalID)

	// TODO: Add tag for if voting period started
	resTags := sdk.NewTags(
		tags.Action, tags.ActionDeposit,
		tags.Depositer, []byte(msg.Depositer.String()),
		tags.ProposalID, proposalIDBytes,
	)

	if votingStarted {
		resTags.AppendTag(tags.VotingPeriodStart, proposalIDBytes)
	}

	return sdk.Result{
		Tags: resTags,
	}
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {

	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(msg.ProposalID)

	resTags := sdk.NewTags(
		tags.Action, tags.ActionVote,
		tags.Voter, []byte(msg.Voter.String()),
		tags.ProposalID, proposalIDBytes,
	)
	return sdk.Result{
		Tags: resTags,
	}
}

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) (resTags sdk.Tags) {

	logger := ctx.Logger().With("module", "x/gov")

	resTags = sdk.NewTags()

	// Delete proposals that haven't met minDeposit
	for shouldPopInactiveProposalQueue(ctx, keeper) {
		inactiveProposal := keeper.InactiveProposalQueuePop(ctx)
		if inactiveProposal.GetStatus() != StatusDepositPeriod {
			continue
		}

		proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(inactiveProposal.GetProposalID())
		keeper.DeleteProposal(ctx, inactiveProposal)
		resTags.AppendTag(tags.Action, tags.ActionProposalDropped)
		resTags.AppendTag(tags.ProposalID, proposalIDBytes)

		logger.Info(fmt.Sprintf("Proposal %d - \"%s\" - didn't mean minimum deposit (had only %s), deleted",
			inactiveProposal.GetProposalID(), inactiveProposal.GetTitle(), inactiveProposal.GetTotalDeposit()))
	}

	// Check if earliest Active Proposal ended voting period yet
	for shouldPopActiveProposalQueue(ctx, keeper) {
		activeProposal := keeper.ActiveProposalQueuePop(ctx)

		proposalStartTime := activeProposal.GetVotingStartTime()
		votingPeriod := keeper.GetVotingProcedure(ctx).VotingPeriod
		if ctx.BlockHeader().Time.Before(proposalStartTime.Add(votingPeriod)) {
			continue
		}

		passes, tallyResults, nonVotingVals := tally(ctx, keeper, activeProposal)
		proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(activeProposal.GetProposalID())
		var action []byte
		if passes {
			keeper.RefundDeposits(ctx, activeProposal.GetProposalID())
			activeProposal.SetStatus(StatusPassed)
			action = tags.ActionProposalPassed
		} else {
			keeper.DeleteDeposits(ctx, activeProposal.GetProposalID())
			activeProposal.SetStatus(StatusRejected)
			action = tags.ActionProposalRejected
		}
		activeProposal.SetTallyResult(tallyResults)
		keeper.SetProposal(ctx, activeProposal)

		logger.Info(fmt.Sprintf("Proposal %d - \"%s\" - tallied, passed: %v",
			activeProposal.GetProposalID(), activeProposal.GetTitle(), passes))

		for _, valAddr := range nonVotingVals {
			val := keeper.ds.GetValidatorSet().Validator(ctx, valAddr)
			keeper.ds.GetValidatorSet().Slash(ctx,
				val.GetPubKey(),
				ctx.BlockHeight(),
				val.GetPower().RoundInt64(),
				keeper.GetTallyingProcedure(ctx).GovernancePenalty)

			logger.Info(fmt.Sprintf("Validator %s failed to vote on proposal %d, slashing",
				val.GetOperator(), activeProposal.GetProposalID()))
		}

		resTags.AppendTag(tags.Action, action)
		resTags.AppendTag(tags.ProposalID, proposalIDBytes)
	}

	return resTags
}
func shouldPopInactiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	depositProcedure := keeper.GetDepositProcedure(ctx)
	peekProposal := keeper.InactiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if peekProposal.GetStatus() != StatusDepositPeriod {
		return true
	} else if !ctx.BlockHeader().Time.Before(peekProposal.GetSubmitTime().Add(depositProcedure.MaxDepositPeriod)) {
		return true
	}
	return false
}

func shouldPopActiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	votingProcedure := keeper.GetVotingProcedure(ctx)
	peekProposal := keeper.ActiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if !ctx.BlockHeader().Time.Before(peekProposal.GetVotingStartTime().Add(votingProcedure.VotingPeriod)) {
		return true
	}
	return false
}
