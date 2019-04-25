package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// Keeper defines the keeper of the minting store
type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *codec.Codec
	paramSpace   params.Subspace
	supplyKeeper SupplyKeeper
	sk           StakingKeeper
	fck          FeeCollectionKeeper
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey,
	paramSpace params.Subspace, supplyKeeper SupplyKeeper, sk StakingKeeper, fck FeeCollectionKeeper) Keeper {

	keeper := Keeper{
		storeKey:     key,
		cdc:          cdc,
		paramSpace:   paramSpace.WithKeyTable(ParamKeyTable()),
		supplyKeeper: supplyKeeper,
		sk:           sk,
		fck:          fck,
	}
	return keeper
}

// get the minter
func (k Keeper) GetMinter(ctx sdk.Context) (minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(minterKey)
	if b == nil {
		panic("Stored minter should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &minter)
	return
}

// set the minter
func (k Keeper) SetMinter(ctx sdk.Context, minter types.Minter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(minter)
	store.Set(minterKey, b)
}

// CalculateInflationRate recalculates the inflation rate and annual provisions for a new block
func (k Keeper) CalculateInflationRate(ctx sdk.Context) {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	bondedRatio := k.sk.BondedRatio(ctx)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, k.sk.StakingTokenSupply(ctx))
	k.SetMinter(ctx, minter)
}

// Mint creates new coins based on the current block provision, which are added
// to the collected fee pool and then updates the total supply
func (k Keeper) Mint(ctx sdk.Context) {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	mintedCoin := minter.BlockProvision(params)
	k.fck.AddCollectedFees(ctx, sdk.NewCoins(mintedCoin))

	// // passively keep track of the total and the not bonded supply
	k.supplyKeeper.InflateSupply(ctx, supply.TypeTotal, sdk.NewCoins(mintedCoin))
	k.sk.InflateNotBondedTokenSupply(ctx, mintedCoin.Amount)
}
