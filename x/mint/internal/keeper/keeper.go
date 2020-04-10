package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper of the mint store
type Keeper struct {
	cdc              *codec.Codec
	storeKey         sdk.StoreKey
	paramSpace       params.Subspace
	sk               types.StakingKeeper
	supplyKeeper     types.SupplyKeeper
	feeCollectorName string
	cache            *types.Cache
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace,
	sk types.StakingKeeper, supplyKeeper types.SupplyKeeper, feeCollectorName string) Keeper {

	// ensure mint module account is set
	if addr := supplyKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         key,
		paramSpace:       paramSpace.WithKeyTable(types.ParamKeyTable()),
		sk:               sk,
		supplyKeeper:     supplyKeeper,
		feeCollectorName: feeCollectorName,
		cache:            &types.Cache{},
	}
}

//______________________________________________________________________

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// get the minter
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b == nil {
		panic("stored minter should not have been nil")
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &minter)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(minter)
	store.Set(types.MinterKey, b)
}

//______________________________________________________________________

// GetParams returns the total set of minting parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	if k.cache.Params != nil {
		return *k.cache.Params
	}
	k.paramSpace.GetParamSet(ctx, &params)
	k.cache.Params = &params
	return params
}

// SetParams sets the total set of minting parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
	k.cache.Params = &params
}

//______________________________________________________________________

// StakingTokenSupply implements an alias call to the underlying staking keeper's
// StakingTokenSupply to be used in BeginBlocker.
func (k Keeper) StakingTokenSupply(ctx sdk.Context) sdk.Dec {
	return k.sk.StakingTokenSupply(ctx)
}

// BondedRatio implements an alias call to the underlying staking keeper's
// BondedRatio to be used in BeginBlocker.
func (k Keeper) BondedRatio(ctx sdk.Context) sdk.Dec {
	return k.sk.BondedRatio(ctx)
}

// MintCoins implements an alias call to the underlying supply keeper's
// MintCoins to be used in BeginBlocker.
func (k Keeper) MintCoins(ctx sdk.Context, newCoins sdk.Coins) sdk.Error {
	if newCoins.Empty() {
		// skip as no coins need to be minted
		return nil
	}
	return k.supplyKeeper.MintCoins(ctx, types.ModuleName, newCoins)
}

// AddCollectedFees implements an alias call to the underlying supply keeper's
// AddCollectedFees to be used in BeginBlocker.
func (k Keeper) AddCollectedFees(ctx sdk.Context, fees sdk.Coins) sdk.Error {
	return k.supplyKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, fees)
}

// get the minter custom
func (k Keeper) GetMinterCustom(ctx sdk.Context) (minter types.MinterCustom) {
	if k.cache.Minter != nil {
		return *k.cache.Minter
	}
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.MinterKey)
	if b != nil {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &minter)
	}

	k.cache.Minter = &minter
	return
}

// set the minter custom
func (k Keeper) SetMinterCustom(ctx sdk.Context, minter types.MinterCustom) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(minter)
	store.Set(types.MinterKey, b)
	k.cache.Minter = &minter
}

func (k Keeper) UpdateMinterCustom(ctx sdk.Context, minter *types.MinterCustom, params types.Params) {
	totalStakingSupply := k.StakingTokenSupply(ctx)
	annualProvisions := params.InflationRate.Mul(totalStakingSupply)
	provisionAmtPerBlock := annualProvisions.Quo(sdk.NewDec(int64(params.BlocksPerYear)))

	// update new MinterCustom
	minter.MintedPerBlock = sdk.NewDecCoinsFromDec(params.MintDenom, provisionAmtPerBlock)
	minter.NextBlockToUpdate += params.BlocksPerYear
	minter.AnnualProvisions = annualProvisions
	k.SetMinterCustom(ctx, *minter)
}
