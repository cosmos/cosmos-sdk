package supply

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper of the supply store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	paramSpace params.Subspace
}

// NewKeeper defines the supply store keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace,
) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
	}
}

// GetTokenHolders returns all the token holders
func (k Keeper) GetTokenHolders(ctx sdk.Context) (
	tokenHolders []TokenHolder, err error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, holderKeyPrefix)
	defer iterator.Close()

	var tokenHolder TokenHolder
	for ; iterator.Valid(); iterator.Next() {
		err = k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &tokenHolder)
		if err != nil {
			return
		}
		tokenHolders = append(tokenHolders, tokenHolder)
	}
	return
}

// GetTokenHolder returns a token holder instance
func (k Keeper) GetTokenHolder(ctx sdk.Context, moduleName string) (
	tokenHolder TokenHolder) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetTokenHolderKey(moduleName))
	if b == nil {
		panic(fmt.Sprintf("module %s is not in store", moduleName))
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, tokenHolder)
	return
}

// SetTokenHolder sets a holder to store
func (k Keeper) SetTokenHolder(ctx sdk.Context, tokenHolder TokenHolder) {
	store := ctx.KVStore(k.storeKey)
	holderKey := GetTokenHolderKey(tokenHolder.module)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(tokenHolder)
	store.Set(holderKey, b)
}

// GetTotalSupply returns the total supply of the network
func (k Keeper) GetTotalSupply(ctx sdk.Context) (totalSupply sdk.Coins) {
	holders, err := k.GetTokenHolders(ctx)
	if err != nil {
		panic(err)
	}

	for _, holder := range holders {
		totalSupply = totalSupply.Add(holder.GetHoldings())
	}
	return
}

// GetSupplyOf returns a coin's total supply
func (k Keeper) GetSupplyOf(ctx sdk.Context, denom string) (supply sdk.Int) {
	holders, err := k.GetTokenHolders(ctx)
	if err != nil {
		panic(err)
	}

	for _, holder := range holders {
		supply = supply.Add(holder.GetHoldingsOf(denom))
	}
	return
}
