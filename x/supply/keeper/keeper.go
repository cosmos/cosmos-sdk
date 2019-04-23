package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// Keeper defines the keeper of the supply store
type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey

	dk  DistributionKeeper
	fck FeeCollectionKeeper
	sk  StakingKeeper
}

// NewKeeper creates a new supply Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey,
	dk DistributionKeeper, fck FeeCollectionKeeper, sk StakingKeeper) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: key,
		dk:       dk,
		fck:      fck,
		sk:       sk,
	}
}

// GetSupplier retrieves the Supplier from store
func (k Keeper) GetSupplier(ctx sdk.Context) (supplier types.Supplier) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(supplierKey)
	if b == nil {
		panic("Stored fee pool should not have been nil")
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &supplier)
	return
}

// SetSupplier sets the Supplier to store
func (k Keeper) SetSupplier(ctx sdk.Context, supplier types.Supplier) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(supplier)
	store.Set(supplierKey, b)
}

// InflateSupply adds tokens to the circulating supply
func (k Keeper) InflateSupply(ctx sdk.Context, supplyType string, amount sdk.Coins) {
	supplier := k.GetSupplier(ctx)
	supplier.Inflate(supplyType, amount)

	k.SetSupplier(ctx, supplier)
}

// TotalSupply returns the total supply of the network
// total = circulating + vesting + modules + bonded supply + collected fees + community pool
func (k Keeper) TotalSupply(ctx sdk.Context) sdk.Coins {
	supplier := k.GetSupplier(ctx)
	supplierTotal := supplier.Total() // circulating + vesting + modules

	bondedSupply := sdk.NewCoins(sdk.NewCoin(k.sk.BondDenom(ctx), k.sk.TotalBondedTokens(ctx)))
	collectedFees := k.fck.GetCollectedFees(ctx)
	communityPool, _ := k.dk.GetFeePoolCommunityCoins(ctx).TruncateDecimal()

	return supplierTotal.Add(bondedSupply).Add(collectedFees).Add(communityPool)
}
