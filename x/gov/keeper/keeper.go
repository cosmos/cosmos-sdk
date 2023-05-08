package keeper

import (
	"context"
	"fmt"
	"time"

	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Keeper defines the governance module Keeper
type Keeper struct {
	authKeeper  types.AccountKeeper
	bankKeeper  types.BankKeeper
	distrKeeper types.DistributionKeeper

	// The reference to the DelegationSet and ValidatorSet to get information about validators and delegators
	sk types.StakingKeeper

	// GovHooks
	hooks types.GovHooks

	// The (unexposed) keys used to access the stores from the Context.
	storeService corestoretypes.KVStoreService

	// The codec for binary encoding/decoding.
	cdc codec.Codec

	// Legacy Proposal router
	legacyRouter v1beta1.Router

	// Msg server router
	router baseapp.MessageRouter

	config types.Config

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// GetAuthority returns the x/gov module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.Codec, storeService corestoretypes.KVStoreService, authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper, sk types.StakingKeeper, distrKeeper types.DistributionKeeper,
	router baseapp.MessageRouter, config types.Config, authority string,
) *Keeper {
	// ensure governance module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if _, err := authKeeper.StringToBytes(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	// If MaxMetadataLen not set by app developer, set to default value.
	if config.MaxMetadataLen == 0 {
		config.MaxMetadataLen = types.DefaultConfig().MaxMetadataLen
	}

	return &Keeper{
		storeService: storeService,
		authKeeper:   authKeeper,
		bankKeeper:   bankKeeper,
		distrKeeper:  distrKeeper,
		sk:           sk,
		cdc:          cdc,
		router:       router,
		config:       config,
		authority:    authority,
	}
}

// Hooks gets the hooks for governance *Keeper {
func (k *Keeper) Hooks() types.GovHooks {
	if k.hooks == nil {
		// return a no-op implementation if no hooks are set
		return types.MultiGovHooks{}
	}

	return k.hooks
}

// SetHooks sets the hooks for governance
func (k *Keeper) SetHooks(gh types.GovHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set governance hooks twice")
	}

	k.hooks = gh

	return k
}

// SetLegacyRouter sets the legacy router for governance
func (k *Keeper) SetLegacyRouter(router v1beta1.Router) {
	// It is vital to seal the governance proposal router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	router.Seal()
	k.legacyRouter = router
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// Router returns the gov keeper's router
func (k Keeper) Router() baseapp.MessageRouter {
	return k.router
}

// LegacyRouter returns the gov keeper's legacy router
func (k Keeper) LegacyRouter() v1beta1.Router {
	return k.legacyRouter
}

// GetGovernanceAccount returns the governance ModuleAccount
func (k Keeper) GetGovernanceAccount(ctx context.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// ProposalQueues

// InsertActiveProposalQueue inserts a proposalID into the active proposal queue at endTime
func (k Keeper) InsertActiveProposalQueue(ctx context.Context, proposalID uint64, endTime time.Time) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := types.GetProposalIDBytes(proposalID)
	return store.Set(types.ActiveProposalQueueKey(proposalID, endTime), bz)
}

// RemoveFromActiveProposalQueue removes a proposalID from the Active Proposal Queue
func (k Keeper) RemoveFromActiveProposalQueue(ctx context.Context, proposalID uint64, endTime time.Time) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.ActiveProposalQueueKey(proposalID, endTime))
}

// InsertInactiveProposalQueue inserts a proposalID into the inactive proposal queue at endTime
func (k Keeper) InsertInactiveProposalQueue(ctx context.Context, proposalID uint64, endTime time.Time) error {
	store := k.storeService.OpenKVStore(ctx)
	bz := types.GetProposalIDBytes(proposalID)
	return store.Set(types.InactiveProposalQueueKey(proposalID, endTime), bz)
}

// RemoveFromInactiveProposalQueue removes a proposalID from the Inactive Proposal Queue
func (k Keeper) RemoveFromInactiveProposalQueue(ctx context.Context, proposalID uint64, endTime time.Time) error {
	store := k.storeService.OpenKVStore(ctx)
	return store.Delete(types.InactiveProposalQueueKey(proposalID, endTime))
}

// Iterators

// IterateActiveProposalsQueue iterates over the proposals in the active proposal queue
// and performs a callback function
func (k Keeper) IterateActiveProposalsQueue(ctx context.Context, endTime time.Time, cb func(proposal v1.Proposal) error) error {
	iterator, err := k.ActiveProposalQueueIterator(ctx, endTime)
	if err != nil {
		return err
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitActiveProposalQueueKey(iterator.Key())
		proposal, err := k.GetProposal(ctx, proposalID)
		if err != nil {
			return err
		}

		err = cb(proposal)
		// exit early without error if cb returns ErrStopIterating
		if errors.IsOf(err, errors.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// IterateInactiveProposalsQueue iterates over the proposals in the inactive proposal queue
// and performs a callback function
func (k Keeper) IterateInactiveProposalsQueue(ctx context.Context, endTime time.Time, cb func(proposal v1.Proposal) error) error {
	iterator, err := k.InactiveProposalQueueIterator(ctx, endTime)
	if err != nil {
		return err
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitInactiveProposalQueueKey(iterator.Key())
		proposal, err := k.GetProposal(ctx, proposalID)
		if err != nil {
			return err
		}

		err = cb(proposal)
		// exit early without error if cb returns ErrStopIterating
		if errors.IsOf(err, errors.ErrStopIterating) {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

// ActiveProposalQueueIterator returns an corestoretypes.Iterator for all the proposals in the Active Queue that expire by endTime
func (k Keeper) ActiveProposalQueueIterator(ctx context.Context, endTime time.Time) (corestoretypes.Iterator, error) {
	store := k.storeService.OpenKVStore(ctx)
	return store.Iterator(types.ActiveProposalQueuePrefix, storetypes.PrefixEndBytes(types.ActiveProposalByTimeKey(endTime)))
}

// InactiveProposalQueueIterator returns an corestoretypes.Iterator for all the proposals in the Inactive Queue that expire by endTime
func (k Keeper) InactiveProposalQueueIterator(ctx context.Context, endTime time.Time) (corestoretypes.Iterator, error) {
	store := k.storeService.OpenKVStore(ctx)
	return store.Iterator(types.InactiveProposalQueuePrefix, storetypes.PrefixEndBytes(types.InactiveProposalByTimeKey(endTime)))
}

// ModuleAccountAddress returns gov module account address
func (k Keeper) ModuleAccountAddress() sdk.AccAddress {
	return k.authKeeper.GetModuleAddress(types.ModuleName)
}

// assertMetadataLength returns an error if given metadata length
// is greater than a pre-defined MaxMetadataLen.
func (k Keeper) assertMetadataLength(metadata string) error {
	if metadata != "" && uint64(len(metadata)) > k.config.MaxMetadataLen {
		return types.ErrMetadataTooLong.Wrapf("got metadata with length %d", len(metadata))
	}
	return nil
}
