package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// keeper of the staking store
type Keeper struct {
	storeKey            sdk.StoreKey
	cdc                 *codec.Codec
	paramSpace          params.Subspace
	bankKeeper          types.BankKeeper
	stakingKeeper       types.StakingKeeper
	feeCollectionKeeper types.FeeCollectionKeeper

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace, ck types.BankKeeper,
	sk types.StakingKeeper, fck types.FeeCollectionKeeper, codespace sdk.CodespaceType) Keeper {

	keeper := Keeper{
		storeKey:            key,
		cdc:                 cdc,
		paramSpace:          paramSpace.WithTypeTable(ParamTypeTable()),
		bankKeeper:          ck,
		stakingKeeper:       sk,
		feeCollectionKeeper: fck,
		codespace:           codespace,
	}
	return keeper
}

//______________________________________________________________________

// get the global fee pool distribution info
func (k Keeper) GetFeePool(ctx sdk.Context) (feePool types.FeePool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(FeePoolKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &feePool)
	return
}

// set the global fee pool distribution info
func (k Keeper) SetFeePool(ctx sdk.Context, feePool types.FeePool) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(feePool)
	store.Set(FeePoolKey, b)
}

// get the total validator accum for the ctx height
// in the fee pool
func (k Keeper) GetFeePoolValAccum(ctx sdk.Context) sdk.Dec {

	// withdraw self-delegation
	height := ctx.BlockHeight()
	totalPower := sdk.NewDecFromInt(k.stakingKeeper.GetLastTotalPower(ctx))
	fp := k.GetFeePool(ctx)
	return fp.GetTotalValAccum(height, totalPower)
}

//______________________________________________________________________

// set the proposer public key for this block
func (k Keeper) GetPreviousProposerConsAddr(ctx sdk.Context) (consAddr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(ProposerKey)
	if b == nil {
		panic("Previous proposer not set")
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &consAddr)
	return
}

// get the proposer public key for this block
func (k Keeper) SetPreviousProposerConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(consAddr)
	store.Set(ProposerKey, b)
}

//______________________________________________________________________

// get context required for withdraw operations
func (k Keeper) GetWithdrawContext(ctx sdk.Context,
	valOperatorAddr sdk.ValAddress) types.WithdrawContext {

	feePool := k.GetFeePool(ctx)
	height := ctx.BlockHeight()
	validator := k.stakingKeeper.Validator(ctx, valOperatorAddr)
	lastValPower := k.stakingKeeper.GetLastValidatorPower(ctx, valOperatorAddr)
	lastTotalPower := sdk.NewDecFromInt(k.stakingKeeper.GetLastTotalPower(ctx))

	return types.NewWithdrawContext(
		feePool, height, lastTotalPower, sdk.NewDecFromInt(lastValPower),
		validator.GetCommission())
}

//______________________________________________________________________
// PARAM STORE

// Type declaration for parameters
func ParamTypeTable() params.TypeTable {
	return params.NewTypeTable(
		ParamStoreKeyCommunityTax, sdk.Dec{},
		ParamStoreKeyBaseProposerReward, sdk.Dec{},
		ParamStoreKeyBonusProposerReward, sdk.Dec{},
	)
}

// Returns the current CommunityTax rate from the global param store
// nolint: errcheck
func (k Keeper) GetCommunityTax(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyCommunityTax, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetCommunityTax(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyCommunityTax, &percent)
}

// Returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetBaseProposerReward(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyBaseProposerReward, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetBaseProposerReward(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyBaseProposerReward, &percent)
}

// Returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetBonusProposerReward(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyBonusProposerReward, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetBonusProposerReward(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyBonusProposerReward, &percent)
}
