package gov

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	// TODO: Find another way to implement this without using accounts, or find a cleaner way to implement it using accounts.
	DepositedCoinsAccAddr     = sdk.AccAddress(crypto.AddressHash([]byte("govDepositedCoins")))
	BurnedDepositCoinsAccAddr = sdk.AccAddress(crypto.AddressHash([]byte("govBurnedDepositCoins")))
)

// Governance Keeper
type Keeper struct {
	// The reference to the Param Keeper to get and set Global Params
	paramsKeeper params.Keeper

	// The reference to the Paramstore to get and set gov specific params
	paramSpace params.Subspace

	// The reference to the CoinKeeper to modify balances
	ck BankKeeper

	// The ValidatorSet to get information about validators
	vs sdk.ValidatorSet

	// The reference to the DelegationSet to get information about delegators
	ds sdk.DelegationSet

	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey

	// The codec codec for binary encoding/decoding.
	cdc *codec.Codec

	// Reserved codespace
	codespace sdk.CodespaceType

	// Proposal router
	router Router
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramsKeeper params.Keeper, paramSpace params.Subspace,
	ck BankKeeper, ds sdk.DelegationSet, codespace sdk.CodespaceType, rtr Router,
) Keeper {

	// It is vital to seal the governance proposal router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	rtr.Seal()

	return Keeper{
		storeKey:     key,
		paramsKeeper: paramsKeeper,
		paramSpace:   paramSpace.WithKeyTable(ParamKeyTable()),
		ck:           ck,
		ds:           ds,
		vs:           ds.GetValidatorSet(),
		cdc:          cdc,
		codespace:    codespace,
		router:       rtr,
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger { return ctx.Logger().With("module", "x/gov") }

// Params

// Returns the current DepositParams from the global param store
func (keeper Keeper) GetDepositParams(ctx sdk.Context) DepositParams {
	var depositParams DepositParams
	keeper.paramSpace.Get(ctx, ParamStoreKeyDepositParams, &depositParams)
	return depositParams
}

// Returns the current VotingParams from the global param store
func (keeper Keeper) GetVotingParams(ctx sdk.Context) VotingParams {
	var votingParams VotingParams
	keeper.paramSpace.Get(ctx, ParamStoreKeyVotingParams, &votingParams)
	return votingParams
}

// Returns the current TallyParam from the global param store
func (keeper Keeper) GetTallyParams(ctx sdk.Context) TallyParams {
	var tallyParams TallyParams
	keeper.paramSpace.Get(ctx, ParamStoreKeyTallyParams, &tallyParams)
	return tallyParams
}

func (keeper Keeper) setDepositParams(ctx sdk.Context, depositParams DepositParams) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyDepositParams, &depositParams)
}

func (keeper Keeper) setVotingParams(ctx sdk.Context, votingParams VotingParams) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyVotingParams, &votingParams)
}

func (keeper Keeper) setTallyParams(ctx sdk.Context, tallyParams TallyParams) {
	keeper.paramSpace.Set(ctx, ParamStoreKeyTallyParams, &tallyParams)
}

// ProposalQueues

// InsertActiveProposalQueue inserts a ProposalID into the active proposal queue at endTime
func (keeper Keeper) InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID)
	store.Set(KeyActiveProposalQueue(proposalID, endTime), bz)
}

// RemoveFromActiveProposalQueue removes a proposalID from the Active Proposal Queue
func (keeper Keeper) RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyActiveProposalQueue(proposalID, endTime))
}

// InsertInactiveProposalQueue Inserts a ProposalID into the inactive proposal queue at endTime
func (keeper Keeper) InsertInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID)
	store.Set(KeyInactiveProposalQueueProposal(endTime, proposalID), bz)
}

// RemoveFromInactiveProposalQueue removes a proposalID from the Inactive Proposal Queue
func (keeper Keeper) RemoveFromInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyInactiveProposalQueue(proposalID, endTime))
}
