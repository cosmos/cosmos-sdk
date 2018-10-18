package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// keeper of the stake store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	paramSpace params.Subspace

	// codespace
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace,
	codespace sdk.CodespaceType) Keeper {

	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		paramSpace: paramSpace.WithTypeTable(ParamTypeTable()),
		codespace:  codespace,
	}
	return keeper
}

//______________________________________________________________________

// get the global fee pool distribution info
func (k Keeper) GetMinter(ctx sdk.Context) (minter Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(minterKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinary(b, &minter)
	return
}

// set the global fee pool distribution info
func (k Keeper) SetMinter(ctx sdk.Context, minter Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinary(minter)
	store.Set(minterKey, b)
}

//______________________________________________________________________

// Returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetInflationRateChange(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyInflationRateChange, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetInflationRateChange(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyInflationRateChange, &percent)
}

// Returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetInflationMax(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyInflationMax, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetInflationMax(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyInflationMax, &percent)
}

// Returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetInflationMin(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyInflationMin, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetInflationMin(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyInflationMin, &percent)
}

// Returns the current BaseProposerReward rate from the global param store
// nolint: errcheck
func (k Keeper) GetGoalBonded(ctx sdk.Context) sdk.Dec {
	var percent sdk.Dec
	k.paramSpace.Get(ctx, ParamStoreKeyGoalBonded, &percent)
	return percent
}

// nolint: errcheck
func (k Keeper) SetGoalBonded(ctx sdk.Context, percent sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyGoalBonded, &percent)
}
