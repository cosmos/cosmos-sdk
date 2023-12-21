package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines the governance module Keeper
type Keeper struct {
	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper
	poolKeeper types.PoolKeeper

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

	Schema       collections.Schema
	Constitution collections.Item[string]
	Params       collections.Item[v1.Params]
	// Deposits key: proposalID+depositorAddr | value: Deposit
	Deposits collections.Map[collections.Pair[uint64, sdk.AccAddress], v1.Deposit]
	// Votes key: proposalID+voterAddr | value: Vote
	Votes      collections.Map[collections.Pair[uint64, sdk.AccAddress], v1.Vote]
	ProposalID collections.Sequence
	// Proposals key:proposalID | value: Proposal
	Proposals collections.Map[uint64, v1.Proposal]
	// ActiveProposalsQueue key: votingEndTime+proposalID | value: proposalID
	ActiveProposalsQueue collections.Map[collections.Pair[time.Time, uint64], uint64] // TODO(tip): this should be simplified and go into an index.
	// InactiveProposalsQueue key: depositEndTime+proposalID | value: proposalID
	InactiveProposalsQueue collections.Map[collections.Pair[time.Time, uint64], uint64] // TODO(tip): this should be simplified and go into an index.
	// VotingPeriodProposals key: proposalID | value: proposalStatus (votingPeriod or not)
	VotingPeriodProposals collections.Map[uint64, []byte] // TODO(tip): this could be a keyset or index.
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
	bankKeeper types.BankKeeper, sk types.StakingKeeper, pk types.PoolKeeper,
	router baseapp.MessageRouter, config types.Config, authority string,
) *Keeper {
	// ensure governance module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if _, err := authKeeper.AddressCodec().StringToBytes(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	defaultConfig := types.DefaultConfig()
	// If MaxMetadataLen not set by app developer, set to default value.
	if config.MaxTitleLen == 0 {
		config.MaxTitleLen = defaultConfig.MaxTitleLen
	}
	// If MaxMetadataLen not set by app developer, set to default value.
	if config.MaxMetadataLen == 0 {
		config.MaxMetadataLen = defaultConfig.MaxMetadataLen
	}
	// If MaxMetadataLen not set by app developer, set to default value.
	if config.MaxSummaryLen == 0 {
		config.MaxSummaryLen = defaultConfig.MaxSummaryLen
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := &Keeper{
		storeService:           storeService,
		authKeeper:             authKeeper,
		bankKeeper:             bankKeeper,
		sk:                     sk,
		poolKeeper:             pk,
		cdc:                    cdc,
		router:                 router,
		config:                 config,
		authority:              authority,
		Constitution:           collections.NewItem(sb, types.ConstitutionKey, "constitution", collections.StringValue),
		Params:                 collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[v1.Params](cdc)),
		Deposits:               collections.NewMap(sb, types.DepositsKeyPrefix, "deposits", collections.PairKeyCodec(collections.Uint64Key, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), codec.CollValue[v1.Deposit](cdc)), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
		Votes:                  collections.NewMap(sb, types.VotesKeyPrefix, "votes", collections.PairKeyCodec(collections.Uint64Key, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), codec.CollValue[v1.Vote](cdc)),          //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
		ProposalID:             collections.NewSequence(sb, types.ProposalIDKey, "proposal_id"),
		Proposals:              collections.NewMap(sb, types.ProposalsKeyPrefix, "proposals", collections.Uint64Key, codec.CollValue[v1.Proposal](cdc)),
		ActiveProposalsQueue:   collections.NewMap(sb, types.ActiveProposalQueuePrefix, "active_proposals_queue", collections.PairKeyCodec(sdk.TimeKey, collections.Uint64Key), collections.Uint64Value),     // sdk.TimeKey is needed to retain state compatibility
		InactiveProposalsQueue: collections.NewMap(sb, types.InactiveProposalQueuePrefix, "inactive_proposals_queue", collections.PairKeyCodec(sdk.TimeKey, collections.Uint64Key), collections.Uint64Value), // sdk.TimeKey is needed to retain state compatibility
		VotingPeriodProposals:  collections.NewMap(sb, types.VotingPeriodProposalKeyPrefix, "voting_period_proposals", collections.Uint64Key, collections.BytesValue),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
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

// ModuleAccountAddress returns gov module account address
func (k Keeper) ModuleAccountAddress() sdk.AccAddress {
	return k.authKeeper.GetModuleAddress(types.ModuleName)
}

// validateProposalLengths checks message metadata, summary and title
// to have the expected length otherwise returns an error.
func (k Keeper) validateProposalLengths(metadata, title, summary string) error {
	if err := k.assertMetadataLength(metadata); err != nil {
		return err
	}
	if err := k.assertSummaryLength(summary); err != nil {
		return err
	}
	if err := k.assertTitleLength(title); err != nil {
		return err
	}
	return nil
}

// assertTitleLength returns an error if given title length
// is greater than a pre-defined MaxTitleLen.
func (k Keeper) assertTitleLength(title string) error {
	if len(title) == 0 {
		return errors.New("proposal title cannot be empty")
	}

	if uint64(len(title)) > k.config.MaxTitleLen {
		return types.ErrTitleTooLong.Wrapf("got title with length %d", len(title))
	}
	return nil
}

// assertMetadataLength returns an error if given metadata length
// is greater than a pre-defined MaxMetadataLen.
func (k Keeper) assertMetadataLength(metadata string) error {
	if uint64(len(metadata)) > k.config.MaxMetadataLen {
		return types.ErrMetadataTooLong.Wrapf("got metadata with length %d", len(metadata))
	}
	return nil
}

// assertSummaryLength returns an error if given summary length
// is greater than a pre-defined MaxSummaryLen.
func (k Keeper) assertSummaryLength(summary string) error {
	if len(summary) == 0 {
		return errors.New("proposal summary cannot be empty")
	}

	if uint64(len(summary)) > k.config.MaxSummaryLen {
		return types.ErrSummaryTooLong.Wrapf("got summary with length %d", len(summary))
	}
	return nil
}
