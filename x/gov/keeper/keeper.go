package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Keeper defines the governance module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set gov specific params
	paramSpace types.ParamSubspace

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	// The reference to the DelegationSet and ValidatorSet to get information about validators and delegators
	sk types.StakingKeeper

	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc codec.BinaryMarshaler

	// Proposal router
	router types.Router
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, paramSpace types.ParamSubspace,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, sk types.StakingKeeper, rtr types.Router,
) Keeper {

	// ensure governance module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// It is vital to seal the governance proposal router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	rtr.Seal()

	return Keeper{
		storeKey:   key,
		paramSpace: paramSpace,
		authKeeper: authKeeper,
		bankKeeper: bankKeeper,
		sk:         sk,
		cdc:        cdc,
		router:     rtr,
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Router returns the gov Keeper's Router
func (keeper Keeper) Router() types.Router {
	return keeper.router
}

// GetGovernanceAccount returns the governance ModuleAccount
func (keeper Keeper) GetGovernanceAccount(ctx sdk.Context) authtypes.ModuleAccountI {
	return keeper.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// ProposalQueues

// InsertActiveProposalQueue inserts a ProposalID into the active proposal queue at endTime
func (keeper Keeper) InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	bz := types.GetProposalIDBytes(proposalID)
	store.Set(types.ActiveProposalQueueKey(proposalID, endTime), bz)
}

// RemoveFromActiveProposalQueue removes a proposalID from the Active Proposal Queue
func (keeper Keeper) RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.ActiveProposalQueueKey(proposalID, endTime))
}

// InsertInactiveProposalQueue Inserts a ProposalID into the inactive proposal queue at endTime
func (keeper Keeper) InsertInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	bz := types.GetProposalIDBytes(proposalID)
	store.Set(types.InactiveProposalQueueKey(proposalID, endTime), bz)
}

// RemoveFromInactiveProposalQueue removes a proposalID from the Inactive Proposal Queue
func (keeper Keeper) RemoveFromInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(types.InactiveProposalQueueKey(proposalID, endTime))
}

// Iterators

// IterateActiveProposalsQueue iterates over the proposals in the active proposal queue
// and performs a callback function
func (keeper Keeper) IterateActiveProposalsQueue(ctx sdk.Context, endTime time.Time, cb func(proposal types.Proposal) (stop bool)) {
	iterator := keeper.ActiveProposalQueueIterator(ctx, endTime)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitActiveProposalQueueKey(iterator.Key())
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}

		if cb(proposal) {
			break
		}
	}
}

// IterateInactiveProposalsQueue iterates over the proposals in the inactive proposal queue
// and performs a callback function
func (keeper Keeper) IterateInactiveProposalsQueue(ctx sdk.Context, endTime time.Time, cb func(proposal types.Proposal) (stop bool)) {
	iterator := keeper.InactiveProposalQueueIterator(ctx, endTime)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitInactiveProposalQueueKey(iterator.Key())
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}

		if cb(proposal) {
			break
		}
	}
}

// ActiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Active Queue that expire by endTime
func (keeper Keeper) ActiveProposalQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(types.ActiveProposalQueuePrefix, sdk.PrefixEndBytes(types.ActiveProposalByTimeKey(endTime)))
}

// InactiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Inactive Queue that expire by endTime
func (keeper Keeper) InactiveProposalQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(types.InactiveProposalQueuePrefix, sdk.PrefixEndBytes(types.InactiveProposalByTimeKey(endTime)))
}
