package gov

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, keeper *keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := keeper.Logger(ctx)

	// delete dead proposals from store and returns theirs deposits.
	// A proposal is dead when it's inactive and didn't get enough deposit on time to get into voting phase.
	keeper.IterateInactiveProposalsQueue(ctx, ctx.BlockHeader().Time, func(proposal v1.Proposal) bool {
		keeper.DeleteProposal(ctx, proposal.Id)
		keeper.RefundAndDeleteDeposits(ctx, proposal.Id) // refund deposit if proposal got removed without getting 100% of the proposal

		// called when proposal become inactive
		keeper.AfterProposalFailedMinDeposit(ctx, proposal.Id)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeInactiveProposal,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute(types.AttributeKeyProposalResult, types.AttributeValueProposalDropped),
			),
		)

		logger.Info(
			"proposal did not meet minimum deposit; deleted",
			"proposal", proposal.Id,
			"min_deposit", sdk.NewCoins(keeper.GetDepositParams(ctx).MinDeposit...).String(),
			"total_deposit", sdk.NewCoins(proposal.TotalDeposit...).String(),
		)

		return false
	})

	// fetch active proposals whose voting periods have ended (are passed the block time)
	keeper.IterateActiveProposalsQueue(ctx, ctx.BlockHeader().Time, func(proposal v1.Proposal) bool {
		var tagValue, logMsg string

		passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)

		if burnDeposits {
			keeper.DeleteAndBurnDeposits(ctx, proposal.Id)
		} else {
			keeper.RefundAndDeleteDeposits(ctx, proposal.Id)
		}

		if passes {
			var (
				idx int
				msg sdk.Msg
			)

			// attempt to execute all messages within the passed proposal
			// Messages may mutate state thus we use a cached context. If one of
			// the handlers fails, no state mutation is written and the error
			// message is logged.
			cacheCtx, writeCache := ctx.CacheContext()
			messages, err := proposal.GetMsgs()
			if err == nil {
				for idx, msg = range messages {
					handler := keeper.Router().Handler(msg)
					_, err = handler(cacheCtx, msg)
					if err != nil {
						break
					}
				}
			}

			// `err == nil` when all handlers passed.
			// Or else, `idx` and `err` are populated with the msg index and error.
			if err == nil {
				proposal.Status = v1.StatusPassed
				tagValue = types.AttributeValueProposalPassed
				logMsg = "passed"

				// The cached context is created with a new EventManager. However, since
				// the proposal handler execution was successful, we want to track/keep
				// any events emitted, so we re-emit to "merge" the events into the
				// original Context's EventManager.
				ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())

				// write state to the underlying multi-store
				writeCache()
			} else {
				proposal.Status = v1.StatusFailed
				tagValue = types.AttributeValueProposalFailed
				logMsg = fmt.Sprintf("passed, but msg %d (%s) failed on execution: %s", idx, sdk.MsgTypeURL(msg), err)
			}
		} else {
			proposal.Status = v1.StatusRejected
			tagValue = types.AttributeValueProposalRejected
			logMsg = "rejected"
		}

		proposal.FinalTallyResult = &tallyResults

		keeper.SetProposal(ctx, proposal)
		keeper.RemoveFromActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)

		// when proposal become active
		keeper.AfterProposalVotingPeriodEnded(ctx, proposal.Id)

		logger.Info(
			"proposal tallied",
			"proposal", proposal.Id,
			"results", logMsg,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeActiveProposal,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute(types.AttributeKeyProposalResult, tagValue),
			),
		)
		return false
	})
}
