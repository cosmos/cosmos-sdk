package gov

import (
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"cosmossdk.io/x/gov/keeper"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called every block.
func EndBlocker(ctx sdk.Context, keeper *keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := keeper.Logger(ctx)
	// delete dead proposals from store and returns theirs deposits.
	// A proposal is dead when it's inactive and didn't get enough deposit on time to get into voting phase.
	rng := collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.HeaderInfo().Time)
	err := keeper.InactiveProposalsQueue.Walk(ctx, rng, func(key collections.Pair[time.Time, uint64], _ uint64) (bool, error) {
		proposal, err := keeper.Proposals.Get(ctx, key.K2())
		if err != nil {
			// if the proposal has an encoding error, this means it cannot be processed by x/gov
			// this could be due to some types missing their registration
			// instead of returning an error (i.e, halting the chain), we fail the proposal
			if errors.Is(err, collections.ErrEncoding) {
				proposal.Id = key.K2()
				if err := failUnsupportedProposal(logger, ctx, keeper, proposal, err.Error(), false); err != nil {
					return false, err
				}

				if err = keeper.DeleteProposal(ctx, proposal.Id); err != nil {
					return false, err
				}

				return false, nil
			}

			return false, err
		}

		if err = keeper.DeleteProposal(ctx, proposal.Id); err != nil {
			return false, err
		}

		params, err := keeper.Params.Get(ctx)
		if err != nil {
			return false, err
		}
		if !params.BurnProposalDepositPrevote {
			err = keeper.RefundAndDeleteDeposits(ctx, proposal.Id) // refund deposit if proposal got removed without getting 100% of the proposal
		} else {
			err = keeper.DeleteAndBurnDeposits(ctx, proposal.Id) // burn the deposit if proposal got removed without getting 100% of the proposal
		}

		if err != nil {
			return false, err
		}

		// called when proposal become inactive
		cacheCtx, writeCache := ctx.CacheContext()
		err = keeper.Hooks().AfterProposalFailedMinDeposit(cacheCtx, proposal.Id)
		if err == nil { // purposely ignoring the error here not to halt the chain if the hook fails
			writeCache()
		} else {
			logger.Error("failed to execute AfterProposalFailedMinDeposit hook", "error", err)
		}

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
			"proposal_type", proposal.ProposalType,
			"title", proposal.Title,
			"min_deposit", sdk.NewCoins(proposal.GetMinDepositFromParams(params)...).String(),
			"total_deposit", sdk.NewCoins(proposal.TotalDeposit...).String(),
		)

		return false, nil
	})
	if err != nil {
		return err
	}

	// fetch active proposals whose voting periods have ended (are passed the block time)
	rng = collections.NewPrefixUntilPairRange[time.Time, uint64](ctx.HeaderInfo().Time)
	err = keeper.ActiveProposalsQueue.Walk(ctx, rng, func(key collections.Pair[time.Time, uint64], _ uint64) (bool, error) {
		proposal, err := keeper.Proposals.Get(ctx, key.K2())
		if err != nil {
			// if the proposal has an encoding error, this means it cannot be processed by x/gov
			// this could be due to some types missing their registration
			// instead of returning an error (i.e, halting the chain), we fail the proposal
			if errors.Is(err, collections.ErrEncoding) {
				proposal.Id = key.K2()
				if err := failUnsupportedProposal(logger, ctx, keeper, proposal, err.Error(), true); err != nil {
					return false, err
				}

				if err = keeper.ActiveProposalsQueue.Remove(ctx, collections.Join(*proposal.VotingEndTime, proposal.Id)); err != nil {
					return false, err
				}

				return false, nil
			}

			return false, err
		}

		var tagValue, logMsg string

		passes, burnDeposits, tallyResults, err := keeper.Tally(ctx, proposal)
		if err != nil {
			return false, err
		}

		// Deposits are always burned if tally said so, regardless of the proposal type.
		// If a proposal passes, deposits are always refunded, regardless of the proposal type.
		// If a proposal fails, and isn't spammy, deposits are refunded, unless the proposal is expedited or optimistic.
		// An expedited or optimistic proposal that fails and isn't spammy is converted to a regular proposal.
		if burnDeposits {
			err = keeper.DeleteAndBurnDeposits(ctx, proposal.Id)
		} else if passes || !(proposal.ProposalType == v1.ProposalType_PROPOSAL_TYPE_EXPEDITED || proposal.ProposalType == v1.ProposalType_PROPOSAL_TYPE_OPTIMISTIC) {
			err = keeper.RefundAndDeleteDeposits(ctx, proposal.Id)
		}
		if err != nil {
			// in case of an error, log it and emit an event
			// we do not want to halt the chain if the refund/burn fails
			// as it could happen due to a governance mistake (governance has let a proposal pass that sends gov funds that were from proposal deposits)

			keeper.Logger(ctx).Error("failed to refund or burn deposits", "error", err)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeProposalDeposit,
					sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
					sdk.NewAttribute(types.AttributeKeyProposalDepositError, "failed to refund or burn deposits"),
					sdk.NewAttribute("error", err.Error()),
				),
			)
		}

		if err = keeper.ActiveProposalsQueue.Remove(ctx, collections.Join(*proposal.VotingEndTime, proposal.Id)); err != nil {
			return false, err
		}

		switch {
		case passes:
			var (
				idx    int
				events sdk.Events
				msg    sdk.Msg
			)

			// attempt to execute all messages within the passed proposal
			// Messages may mutate state thus we use a cached context. If one of
			// the handlers fails, no state mutation is written and the error
			// message is logged.
			cacheCtx, writeCache := ctx.CacheContext()
			messages, err := proposal.GetMsgs()
			if err != nil {
				proposal.Status = v1.StatusFailed
				proposal.FailedReason = err.Error()
				tagValue = types.AttributeValueProposalFailed
				logMsg = fmt.Sprintf("passed proposal (%v) failed to execute; msgs: %s", proposal, err)

				break
			}

			// execute all messages
			for idx, msg = range messages {
				handler := keeper.Router().Handler(msg)
				var res *sdk.Result
				res, err = safeExecuteHandler(cacheCtx, msg, handler)
				if err != nil {
					break
				}

				events = append(events, res.GetEvents()...)
			}

			// `err == nil` when all handlers passed.
			// Or else, `idx` and `err` are populated with the msg index and error.
			if err == nil {
				proposal.Status = v1.StatusPassed
				tagValue = types.AttributeValueProposalPassed
				logMsg = "passed"

				// write state to the underlying multi-store
				writeCache()

				// propagate the msg events to the current context
				ctx.EventManager().EmitEvents(events)
			} else {
				proposal.Status = v1.StatusFailed
				proposal.FailedReason = err.Error()
				tagValue = types.AttributeValueProposalFailed
				logMsg = fmt.Sprintf("passed, but msg %d (%s) failed on execution: %s", idx, sdk.MsgTypeURL(msg), err)
			}
		case !burnDeposits && (proposal.ProposalType == v1.ProposalType_PROPOSAL_TYPE_EXPEDITED ||
			proposal.ProposalType == v1.ProposalType_PROPOSAL_TYPE_OPTIMISTIC):
			// When a non spammy expedited/optimistic proposal fails, it is converted
			// to a regular proposal. As a result, the voting period is extended, and,
			// once the regular voting period expires again, the tally is repeated
			// according to the regular proposal rules.
			proposal.ProposalType = v1.ProposalType_PROPOSAL_TYPE_STANDARD
			proposal.Expedited = false // can be removed as never read but kept for state coherence
			params, err := keeper.Params.Get(ctx)
			if err != nil {
				return false, err
			}
			endTime := proposal.VotingStartTime.Add(*params.VotingPeriod)
			proposal.VotingEndTime = &endTime

			err = keeper.ActiveProposalsQueue.Set(ctx, collections.Join(*proposal.VotingEndTime, proposal.Id), proposal.Id)
			if err != nil {
				return false, err
			}

			if proposal.ProposalType == v1.ProposalType_PROPOSAL_TYPE_EXPEDITED {
				tagValue = types.AttributeValueExpeditedProposalRejected
				logMsg = "expedited proposal converted to regular"
			} else {
				tagValue = types.AttributeValueOptimisticProposalRejected
				logMsg = "optimistic proposal converted to regular"
			}
		default:
			proposal.Status = v1.StatusRejected
			proposal.FailedReason = "proposal did not get enough votes to pass"
			tagValue = types.AttributeValueProposalRejected
			logMsg = "rejected"
		}

		proposal.FinalTallyResult = &tallyResults

		err = keeper.SetProposal(ctx, proposal)
		if err != nil {
			return false, err
		}

		// when proposal become active
		cacheCtx, writeCache := ctx.CacheContext()
		err = keeper.Hooks().AfterProposalVotingPeriodEnded(cacheCtx, proposal.Id)
		if err == nil { // purposely ignoring the error here not to halt the chain if the hook fails
			writeCache()
		} else {
			logger.Error("failed to execute AfterProposalVotingPeriodEnded hook", "error", err)
		}

		logger.Info(
			"proposal tallied",
			"proposal", proposal.Id,
			"proposal_type", proposal.ProposalType,
			"status", proposal.Status.String(),
			"title", proposal.Title,
			"results", logMsg,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeActiveProposal,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute(types.AttributeKeyProposalResult, tagValue),
				sdk.NewAttribute(types.AttributeKeyProposalLog, logMsg),
			),
		)

		return false, nil
	})
	return err
}

// executes handle(msg) and recovers from panic.
func safeExecuteHandler(ctx sdk.Context, msg sdk.Msg, handler baseapp.MsgServiceHandler,
) (res *sdk.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handling x/gov proposal msg [%s] PANICKED: %v", msg, r)
		}
	}()
	res, err = handler(ctx, msg)
	return
}

// failUnsupportedProposal fails a proposal that cannot be processed by gov
func failUnsupportedProposal(
	logger log.Logger,
	ctx sdk.Context,
	keeper *keeper.Keeper,
	proposal v1.Proposal,
	errMsg string,
	active bool,
) error {
	proposal.Status = v1.StatusFailed
	proposal.FailedReason = fmt.Sprintf("proposal failed because it cannot be processed by gov: %s", errMsg)
	proposal.Messages = nil // clear out the messages

	if err := keeper.SetProposal(ctx, proposal); err != nil {
		return err
	}

	if err := keeper.RefundAndDeleteDeposits(ctx, proposal.Id); err != nil {
		return err
	}

	eventType := types.EventTypeInactiveProposal
	if active {
		eventType = types.EventTypeActiveProposal
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
			sdk.NewAttribute(types.AttributeKeyProposalResult, types.AttributeValueProposalFailed),
		),
	)

	logger.Info(
		"proposal failed to decode; deleted",
		"proposal", proposal.Id,
		"proposal_type", proposal.ProposalType,
		"title", proposal.Title,
		"results", errMsg,
	)

	return nil
}
