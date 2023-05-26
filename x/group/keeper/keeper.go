package keeper

import (
	"fmt"
	"time"

	groupv1 "cosmossdk.io/api/cosmos/group/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/orm/model/ormdb"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

type Keeper struct {
	accKeeper group.AccountKeeper
	state     groupv1.StateStore
	router    baseapp.MessageRouter
	config    group.Config
	modDb     ormdb.ModuleDB

	// used for migrations
	storeService store.KVStoreService
	cdc          codec.Codec
}

// NewKeeper creates a new group keeper.
func NewKeeper(storeService store.KVStoreService, cdc codec.Codec, router baseapp.MessageRouter, accKeeper group.AccountKeeper, config group.Config) Keeper {
	modDb, err := ormdb.NewModuleDB(group.ORMSchema, ormdb.ModuleDBOptions{KVStoreService: storeService})
	if err != nil {
		panic(err)
	}

	state, err := groupv1.NewStateStore(modDb)
	if err != nil {
		panic(err)
	}

	if config.MaxMetadataLen == 0 {
		config.MaxMetadataLen = group.DefaultConfig().MaxMetadataLen
	}
	if config.MaxExecutionPeriod == 0 {
		config.MaxExecutionPeriod = group.DefaultConfig().MaxExecutionPeriod
	}

	return Keeper{
		router:    router,
		accKeeper: accKeeper,
		state:     state,
		config:    config,
		modDb:     modDb,
		// used for migrations
		storeService: storeService,
		cdc:          cdc,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", group.ModuleName))
}

// GenesisHandler returns the genesis handler for the group module.
func (k Keeper) GenesisHandler() appmodule.HasGenesis {
	return k.modDb.GenesisHandler()
}

// proposalsByVPEnd returns all proposals whose voting_period_end is after the `endTime` time argument.
func (k Keeper) proposalsByVPEnd(ctx sdk.Context, endTime time.Time) (proposals []group.Proposal, err error) {
	it, err := k.state.ProposalTable().List(ctx, groupv1.ProposalVotingPeriodEndIndexKey{}.WithVotingPeriodEnd(timestamppb.New(endTime)))
	if err != nil {
		return proposals, err
	}
	defer it.Close()

	for it.Next() {
		// Important: this following line cannot be outside of the for loop.
		// It seems that when one unmarshals into the same `group.Proposal`
		// reference, then gogoproto somehow "adds" the new bytes to the old
		// object for some fields. When running simulations, for proposals with
		// each 1-2 proposers, after a couple of loop iterations we got to a
		// proposal with 60k+ proposers.
		// So we're declaring a local variable that gets GCed.
		//
		// Also see `x/group/types/proposal_test.go`, TestGogoUnmarshalProposal().
		proposal, err := it.Value()
		if err != nil {
			return proposals, err
		}

		proposals = append(proposals, group.ProposalFromPulsar(k.cdc, proposal))
	}

	return proposals, nil
}

// pruneProposal deletes a proposal from state.
func (k Keeper) pruneProposal(ctx sdk.Context, proposalID uint64) error {
	if err := k.state.ProposalTable().Delete(ctx, &groupv1.Proposal{Id: proposalID}); err != nil {
		return err
	}

	k.Logger(ctx).Debug("pruned proposal", "id", proposalID)
	return nil
}

// abortProposals iterates through all proposals by group policy index
// and marks submitted proposals as aborted.
func (k Keeper) abortProposals(ctx sdk.Context, groupPolicyAddr sdk.AccAddress) error {
	proposals, err := k.proposalsByGroupPolicy(ctx, groupPolicyAddr)
	if err != nil {
		return err
	}

	for _, proposalInfo := range proposals {
		// Mark all proposals still in the voting phase as aborted.
		if proposalInfo.Status == groupv1.ProposalStatus_PROPOSAL_STATUS_SUBMITTED {
			proposalInfo.Status = groupv1.ProposalStatus_PROPOSAL_STATUS_ABORTED

			if err := k.state.ProposalTable().Update(ctx, proposalInfo); err != nil {
				return err
			}
		}
	}
	return nil
}

// proposalsByGroupPolicy returns all proposals for a given group policy.
func (k Keeper) proposalsByGroupPolicy(ctx sdk.Context, groupPolicyAddr sdk.AccAddress) ([]*groupv1.Proposal, error) {
	it, err := k.state.ProposalTable().List(ctx, groupv1.ProposalGroupPolicyAddressIndexKey{}.WithGroupPolicyAddress(groupPolicyAddr.String()))
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var proposals []*groupv1.Proposal
	for it.Next() {
		proposalInfo, err := it.Value()
		if err != nil {
			return proposals, err
		}

		proposals = append(proposals, proposalInfo)
	}
	return proposals, nil
}

// pruneVotes prunes all votes for a proposal from state.
func (k Keeper) pruneVotes(ctx sdk.Context, proposalID uint64) error {
	votes, err := k.votesByProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	for _, v := range votes {
		v := v

		if err := k.state.VoteTable().Delete(ctx, &groupv1.Vote{
			ProposalId: proposalID,
			Voter:      v.Voter,
		}); err != nil {
			return err
		}
	}

	return nil
}

// votesByProposal returns all votes for a given proposal.
func (k Keeper) votesByProposal(ctx sdk.Context, proposalID uint64) ([]*groupv1.Vote, error) {
	it, err := k.state.VoteTable().List(ctx, groupv1.VoteProposalIdVoterIndexKey{}.WithProposalId(proposalID))
	if err != nil {
		return nil, err
	}
	defer it.Close()

	var votes []*groupv1.Vote
	for it.Next() {
		vote, err := it.Value()
		if err != nil {
			return votes, err
		}

		votes = append(votes, vote)
	}

	return votes, nil
}

// GetGroupSequence returns the current value of the group table sequence
func (k Keeper) GetGroupSequence(ctx sdk.Context) (uint64, error) {
	return k.state.GroupInfoTable().LastInsertedSequence(ctx)
}

// PruneProposals prunes all proposals that are expired, i.e. whose
// `voting_period + max_execution_period` is greater than the current block
// time.
func (k Keeper) PruneProposals(ctx sdk.Context) error {
	proposals, err := k.proposalsByVPEnd(ctx, ctx.BlockTime().Add(-k.config.MaxExecutionPeriod))
	if err != nil {
		return nil
	}
	for _, proposal := range proposals {
		err := k.pruneProposal(ctx, proposal.Id)
		if err != nil {
			return err
		}
		// Emit event for proposal finalized with its result
		if err := ctx.EventManager().EmitTypedEvent(
			&group.EventProposalPruned{
				ProposalId:  proposal.Id,
				Status:      proposal.Status,
				TallyResult: &proposal.FinalTallyResult,
			}); err != nil {
			return err
		}
	}

	return nil
}

// TallyProposalsAtVPEnd iterates over all proposals whose voting period
// has ended, tallies their votes, prunes them, and updates the proposal's
// `FinalTallyResult` field.
func (k Keeper) TallyProposalsAtVPEnd(ctx sdk.Context) error {
	proposals, err := k.proposalsByVPEnd(ctx, ctx.BlockTime())
	if err != nil {
		return nil
	}

	for _, proposal := range proposals {
		proposal := proposal

		policyInfo, err := k.getGroupPolicyInfo(ctx, proposal.GroupPolicyAddress)
		if err != nil {
			return errorsmod.Wrap(err, "group policy")
		}

		electorate, err := k.getGroupInfo(ctx, policyInfo.GroupId)
		if err != nil {
			return errorsmod.Wrap(err, "group")
		}

		proposalID := proposal.Id
		if proposal.Status == group.PROPOSAL_STATUS_ABORTED || proposal.Status == group.PROPOSAL_STATUS_WITHDRAWN {
			if err := k.pruneProposal(ctx, proposalID); err != nil {
				return err
			}
			if err := k.pruneVotes(ctx, proposalID); err != nil {
				return err
			}
			// Emit event for proposal finalized with its result
			if err := ctx.EventManager().EmitTypedEvent(
				&group.EventProposalPruned{
					ProposalId: proposal.Id,
					Status:     proposal.Status,
				}); err != nil {
				return err
			}
		} else if proposal.Status == group.PROPOSAL_STATUS_SUBMITTED {
			if err := k.doTallyAndUpdate(ctx, &proposal, electorate, policyInfo); err != nil {
				return errorsmod.Wrap(err, "doTallyAndUpdate")
			}

			if err := k.state.ProposalTable().Update(ctx, group.ProposalToPulsar(proposal)); err != nil {
				return errorsmod.Wrap(err, "proposal update")
			}
		}
		// Note: We do nothing if the proposal has been marked as ACCEPTED or
		// REJECTED.
	}
	return nil
}
