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

	id, err := keeper.NewTextProposal(ctx, msg.Title, msg.Description, msg.ProposalType)
	if err != nil {
		panic(err)
	}

	sdkerr, votingStarted := keeper.AddDeposit(ctx, id, msg.Proposer, msg.InitialDeposit)
	if sdkerr != nil {
		return sdkerr.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(id)

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
	for shouldPopInactiveInfoQueue(ctx, keeper) {
		inactiveProposalInfo := keeper.InactiveInfoQueuePop(ctx)
		if inactiveProposalInfo.Status != StatusDepositPeriod {
			continue
		}

		proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(inactiveProposalInfo.ProposalID)
		keeper.DeleteProposalInfo(ctx, inactiveProposalInfo.ProposalID)
		resTags.AppendTag(tags.Action, tags.ActionProposalDropped)
		resTags.AppendTag(tags.ProposalID, proposalIDBytes)

		logger.Info(
			fmt.Sprintf("proposal %d didn't meet minimum deposit of %v steak (had only %v steak); deleted",
				inactiveProposalInfo.ProposalID,
				keeper.GetDepositProcedure(ctx).MinDeposit.AmountOf("steak"),
				inactiveProposalInfo.TotalDeposit.AmountOf("steak"),
			),
		)
	}

	// Check if earliest Active Proposal ended voting period yet
	for shouldPopActiveInfoQueue(ctx, keeper) {
		activeProposalInfo := keeper.ActiveInfoQueuePop(ctx)

		proposalStartTime := activeProposalInfo.VotingStartTime
		votingPeriod := keeper.GetVotingProcedure(ctx).VotingPeriod
		if ctx.BlockHeader().Time.Before(proposalStartTime.Add(votingPeriod)) {
			continue
		}

		proposalID := activeProposalInfo.ProposalID
		activeProposal := keeper.GetProposal(ctx, proposalID)
		abstract := activeProposal.GetProposalAbstract()
		passes, tallyResults := tally(ctx, keeper, proposalID)
		proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(proposalID)
		var action []byte
		if passes {
			keeper.RefundDeposits(ctx, proposalID)
			activeProposalInfo.Status = StatusPassed
			action = tags.ActionProposalPassed
			err := activeProposal.Enact(ctx, keeper)
			if err != nil {
				logger.Info(fmt.Sprintf("proposal %d (%s) returned error while being enacted; error msg: %s",
					proposalID, abstract.Title, err.Error()))
				action = tags.ActionProposalError
			}
		} else {
			keeper.DeleteDeposits(ctx, proposalID)
			activeProposalInfo.Status = StatusRejected
			action = tags.ActionProposalRejected
		}
		activeProposalInfo.TallyResult = tallyResults
		keeper.SetProposalInfo(ctx, activeProposalInfo)

		logger.Info(fmt.Sprintf("proposal %d (%s) tallied; passed: %v",
			proposalID, abstract.Title, passes))

		resTags.AppendTag(tags.Action, action)
		resTags.AppendTag(tags.ProposalID, proposalIDBytes)
	}

	return resTags
}
func shouldPopInactiveInfoQueue(ctx sdk.Context, keeper Keeper) bool {
	depositProcedure := keeper.GetDepositProcedure(ctx)
	peekProposal := keeper.InactiveInfoQueuePeek(ctx)

	if peekProposal.ProposalID == 0 {
		return false
	} else if peekProposal.Status != StatusDepositPeriod {
		return true
	} else if !ctx.BlockHeader().Time.Before(peekProposal.SubmitTime.Add(depositProcedure.MaxDepositPeriod)) {
		return true
	}
	return false
}

func shouldPopActiveInfoQueue(ctx sdk.Context, keeper Keeper) bool {
	votingProcedure := keeper.GetVotingProcedure(ctx)
	peekProposal := keeper.ActiveInfoQueuePeek(ctx)

	if peekProposal.ProposalID == 0 {
		return false
	} else if !ctx.BlockHeader().Time.Before(peekProposal.VotingStartTime.Add(votingProcedure.VotingPeriod)) {
		return true
	}
	return false
}
