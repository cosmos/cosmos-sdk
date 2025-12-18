package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

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

	Schema                                collections.Schema
	Constitution                          collections.Item[string]
	Params                                collections.Item[v1.Params]
	Deposits                              collections.Map[collections.Pair[uint64, sdk.AccAddress], v1.Deposit]
	Votes                                 collections.Map[collections.Pair[uint64, sdk.AccAddress], v1.Vote]
	ProposalID                            collections.Sequence
	Proposals                             collections.Map[uint64, v1.Proposal]
	ActiveProposalsQueue                  collections.Map[collections.Pair[time.Time, uint64], uint64] // TODO(tip): this should be simplified and go into an index.
	InactiveProposalsQueue                collections.Map[collections.Pair[time.Time, uint64], uint64] // TODO(tip): this should be simplified and go into an index.
	VotingPeriodProposals                 collections.Map[uint64, []byte]                              // TODO(tip): this could be a keyset or index.
	LastMinDeposit                        collections.Item[v1.LastMinDeposit]
	LastMinInitialDeposit                 collections.Item[v1.LastMinDeposit]
	ParticipationEMA                      collections.Item[math.LegacyDec]
	ConstitutionAmendmentParticipationEMA collections.Item[math.LegacyDec]
	LawParticipationEMA                   collections.Item[math.LegacyDec]
	QuorumCheckQueue                      collections.Map[collections.Pair[time.Time, uint64], v1.QuorumCheckQueueEntry]
	Governors                             collections.Map[types.GovernorAddress, v1.Governor]
	GovernanceDelegations                 collections.Map[sdk.AccAddress, v1.GovernanceDelegation]
	GovernanceDelegationsByGovernor       collections.Map[collections.Pair[types.GovernorAddress, sdk.AccAddress], v1.GovernanceDelegation]
	ValidatorSharesByGovernor             collections.Map[collections.Pair[types.GovernorAddress, sdk.ValAddress], v1.GovernorValShares]
}

// GetAuthority returns the x/gov module's authority.
func (keeper Keeper) GetAuthority() string {
	return keeper.authority
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

	if _, err := authKeeper.AddressCodec().StringToBytes(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	// If MaxMetadataLen not set by app developer, set to default value.
	if config.MaxMetadataLen == 0 {
		config.MaxMetadataLen = types.DefaultConfig().MaxMetadataLen
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := &Keeper{
		storeService:                          storeService,
		authKeeper:                            authKeeper,
		bankKeeper:                            bankKeeper,
		distrKeeper:                           distrKeeper,
		sk:                                    sk,
		cdc:                                   cdc,
		router:                                router,
		config:                                config,
		authority:                             authority,
		Constitution:                          collections.NewItem(sb, types.ConstitutionKey, "constitution", collections.StringValue),
		Params:                                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[v1.Params](cdc)),
		Deposits:                              collections.NewMap(sb, types.DepositsKeyPrefix, "deposits", collections.PairKeyCodec(collections.Uint64Key, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), codec.CollValue[v1.Deposit](cdc)), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
		Votes:                                 collections.NewMap(sb, types.VotesKeyPrefix, "votes", collections.PairKeyCodec(collections.Uint64Key, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), codec.CollValue[v1.Vote](cdc)),          // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
		ProposalID:                            collections.NewSequence(sb, types.ProposalIDKey, "proposal_id"),
		Proposals:                             collections.NewMap(sb, types.ProposalsKeyPrefix, "proposals", collections.Uint64Key, codec.CollValue[v1.Proposal](cdc)),
		ActiveProposalsQueue:                  collections.NewMap(sb, types.ActiveProposalQueuePrefix, "active_proposals_queue", collections.PairKeyCodec(sdk.TimeKey, collections.Uint64Key), collections.Uint64Value),     // sdk.TimeKey is needed to retain state compatibility
		InactiveProposalsQueue:                collections.NewMap(sb, types.InactiveProposalQueuePrefix, "inactive_proposals_queue", collections.PairKeyCodec(sdk.TimeKey, collections.Uint64Key), collections.Uint64Value), // sdk.TimeKey is needed to retain state compatibility
		VotingPeriodProposals:                 collections.NewMap(sb, types.VotingPeriodProposalKeyPrefix, "voting_period_proposals", collections.Uint64Key, collections.BytesValue),
		LastMinDeposit:                        collections.NewItem(sb, types.LastMinDepositKey, "last_min_deposit", codec.CollValue[v1.LastMinDeposit](cdc)),
		LastMinInitialDeposit:                 collections.NewItem(sb, types.LastMinInitialDepositKey, "last_min_initial_deposit", codec.CollValue[v1.LastMinDeposit](cdc)),
		ParticipationEMA:                      collections.NewItem(sb, types.ParticipationEMAKey, "participation_ema", sdk.LegacyDecValue),
		LawParticipationEMA:                   collections.NewItem(sb, types.LawParticipationEMAKey, "law_participation_ema", sdk.LegacyDecValue),
		ConstitutionAmendmentParticipationEMA: collections.NewItem(sb, types.ConstitutionAmendmentParticipationEMAKey, "constitution_amendment_participation_ema", sdk.LegacyDecValue),
		QuorumCheckQueue:                      collections.NewMap(sb, types.QuorumCheckQueuePrefix, "quorum_check_queue", collections.PairKeyCodec(sdk.TimeKey, collections.Uint64Key), codec.CollValue[v1.QuorumCheckQueueEntry](cdc)),
		Governors:                             collections.NewMap(sb, types.GovernorsKeyPrefix, "governors", types.GovernorAddressKey, codec.CollValue[v1.Governor](cdc)),
		GovernanceDelegations:                 collections.NewMap(sb, types.GovernanceDelegationKeyPrefix, "governance_delegations", sdk.AccAddressKey, codec.CollValue[v1.GovernanceDelegation](cdc)),
		GovernanceDelegationsByGovernor:       collections.NewMap(sb, types.GovernanceDelegationsByGovernorKeyPrefix, "governance_delegations_by_governor", collections.PairKeyCodec(types.GovernorAddressKey, sdk.AccAddressKey), codec.CollValue[v1.GovernanceDelegation](cdc)),
		ValidatorSharesByGovernor:             collections.NewMap(sb, types.ValidatorSharesByGovernorKeyPrefix, "validator_shares_by_governor", collections.PairKeyCodec(types.GovernorAddressKey, sdk.ValAddressKey), codec.CollValue[v1.GovernorValShares](cdc)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Hooks gets the hooks for governance *Keeper {
func (keeper *Keeper) Hooks() types.GovHooks {
	if keeper.hooks == nil {
		// return a no-op implementation if no hooks are set
		return types.MultiGovHooks{}
	}

	return keeper.hooks
}

// SetHooks sets the hooks for governance
func (keeper *Keeper) SetHooks(gh types.GovHooks) *Keeper {
	if keeper.hooks != nil {
		panic("cannot set governance hooks twice")
	}

	keeper.hooks = gh

	return keeper
}

// SetLegacyRouter sets the legacy router for governance
func (keeper *Keeper) SetLegacyRouter(router v1beta1.Router) {
	// It is vital to seal the governance proposal router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	router.Seal()
	keeper.legacyRouter = router
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// Router returns the gov keeper's router
func (keeper Keeper) Router() baseapp.MessageRouter {
	return keeper.router
}

// LegacyRouter returns the gov keeper's legacy router
func (keeper Keeper) LegacyRouter() v1beta1.Router {
	return keeper.legacyRouter
}

// GetGovernanceAccount returns the governance ModuleAccount
func (keeper Keeper) GetGovernanceAccount(ctx context.Context) sdk.ModuleAccountI {
	return keeper.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// ModuleAccountAddress returns gov module account address
func (keeper Keeper) ModuleAccountAddress() sdk.AccAddress {
	return keeper.authKeeper.GetModuleAddress(types.ModuleName)
}

// assertMetadataLength returns an error if given metadata length
// is greater than a pre-defined MaxMetadataLen.
func (keeper Keeper) assertMetadataLength(metadata string) error {
	if metadata != "" && uint64(len(metadata)) > keeper.config.MaxMetadataLen {
		return types.ErrMetadataTooLong.Wrapf("got metadata with length %d", len(metadata))
	}
	return nil
}

// assertSummaryLength returns an error if given summary length
// is greater than a pre-defined 40*MaxMetadataLen.
func (keeper Keeper) assertSummaryLength(summary string) error {
	if summary != "" && uint64(len(summary)) > 40*keeper.config.MaxMetadataLen {
		return types.ErrSummaryTooLong.Wrapf("got summary with length %d", len(summary))
	}
	return nil
}
