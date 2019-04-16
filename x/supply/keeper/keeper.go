package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// Keeper defines the keeper of the supply store
type Keeper struct {
	cdc      *codec.Codec
	storeKey sdk.StoreKey
}

// NewKeeper creates a new supply Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: key,
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
func (k Keeper) InflateSupply(ctx sdk.Context, amount sdk.Coins) {
	supplier := k.GetSupplier(ctx)
	supplier.CirculatingSupply = supplier.CirculatingSupply.Add(amount)
	k.SetSupplier(ctx, supplier)
}

// GetTokenHolders returns all the token holders
func (k Keeper) GetTokenHolders(ctx sdk.Context) (
	tokenHolders []types.TokenHolder) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, holderKeyPrefix)
	defer iterator.Close()

	var tokenHolder types.TokenHolder
	for ; iterator.Valid(); iterator.Next() {
		err := k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &tokenHolder)
		if err != nil {
			panic(err)
		}
		tokenHolders = append(tokenHolders, tokenHolder)
	}
	return
}

// GetTokenHolder returns a token holder instance
func (k Keeper) GetTokenHolder(ctx sdk.Context, moduleName string) (
	tokenHolder types.TokenHolder, err error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(GetTokenHolderKey(moduleName))
	if b == nil {
		err = fmt.Errorf("token holder with module %s doesn't exist", moduleName)
		return
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, tokenHolder)
	return
}

// AddTokenHolder creates and sets a token holder instance to store
func (k Keeper) AddTokenHolder(ctx sdk.Context, moduleName string) (
	tokenHolder types.TokenHolder, err sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	if store.Has(GetTokenHolderKey(moduleName)) {
		err = types.ErrInvalidTokenHolder(types.DefaultCodespace,
			fmt.Sprintf("token holder with module %s already exist", moduleName),
		)
		return
	}

	tokenHolder = types.NewBaseTokenHolder(moduleName, sdk.Coins{})
	k.SetTokenHolder(ctx, tokenHolder)
	return
}

// SetTokenHolder sets a holder to store
func (k Keeper) SetTokenHolder(ctx sdk.Context, tokenHolder types.TokenHolder) {
	store := ctx.KVStore(k.storeKey)
	holderKey := GetTokenHolderKey(tokenHolder.GetModuleName())
	b := k.cdc.MustMarshalBinaryLengthPrefixed(tokenHolder)
	store.Set(holderKey, b)
}

// RequestTokens adds requested tokens to the module's holdings
func (k Keeper) RequestTokens(
	ctx sdk.Context, moduleName string, amount sdk.Coins,
) sdk.Error {
	if !amount.IsValid() {
		return sdk.ErrInvalidCoins("invalid requested amount")
	}

	holder, err := k.GetTokenHolder(ctx, moduleName)
	if err != nil {
		return types.ErrUnknownTokenHolder(
			types.DefaultCodespace,
			fmt.Sprintf("token holder %s doesn't exist", moduleName),
		)
	}

	// update global supply held by token holders
	supplier := k.GetSupplier(ctx)
	supplier.HoldersSupply = supplier.HoldersSupply.Add(amount)

	holder.SetHoldings(holder.GetHoldings().Add(amount))

	k.SetTokenHolder(ctx, holder)
	k.SetSupplier(ctx, supplier)
	return nil
}

// RelinquishTokens hands over a portion of the module's holdings
func (k Keeper) RelinquishTokens(
	ctx sdk.Context, moduleName string, amount sdk.Coins,
) sdk.Error {
	if !amount.IsValid() {
		return sdk.ErrInvalidCoins("invalid provided relenquished amount")
	}

	holder, err := k.GetTokenHolder(ctx, moduleName)
	if err != nil {
		return types.ErrUnknownTokenHolder(
			types.DefaultCodespace,
			fmt.Sprintf("token holder %s doesn't exist", moduleName),
		)
	}

	newHoldings, ok := holder.GetHoldings().SafeSub(amount)
	if !ok {
		return sdk.ErrInsufficientCoins("insufficient token holdings")
	}

	// update global supply held by token holders
	supplier := k.GetSupplier(ctx)
	newHoldersSupply, ok := supplier.HoldersSupply.SafeSub(amount)
	if !ok {
		panic("total holders supply should be greater than relinquished amount")
	}
	supplier.HoldersSupply = newHoldersSupply

	holder.SetHoldings(newHoldings)

	k.SetTokenHolder(ctx, holder)
	k.SetSupplier(ctx, supplier)
	return nil
}
