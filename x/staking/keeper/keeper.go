package keeper

import (
	"container/list"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const aminoCacheSize = 500

// names used as root for pool module accounts:
//
// - NotBondedPool -> "NotBondedTokens"
//
// - BondedPool -> "BondedTokens"
const (
	NotBondedTokensName = "NotBondedTokens"
	BondedTokensName    = "BondedTokens"
)

// keeper of the staking store
type Keeper struct {
	storeKey           sdk.StoreKey
	storeTKey          sdk.StoreKey
	cdc                *codec.Codec
	bankKeeper         types.BankKeeper
	supplyKeeper       types.SupplyKeeper
	hooks              sdk.StakingHooks
	paramstore         params.Subspace
	validatorCache     map[string]cachedValidator
	validatorCacheList *list.List

	// codespace
	codespace sdk.CodespaceType
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(cdc *codec.Codec, key, tkey sdk.StoreKey, bk types.BankKeeper,
	sk types.SupplyKeeper, paramstore params.Subspace, codespace sdk.CodespaceType) Keeper {

	return Keeper{
		storeKey:           key,
		storeTKey:          tkey,
		cdc:                cdc,
		bankKeeper:         bk,
		supplyKeeper:       sk,
		paramstore:         paramstore.WithKeyTable(ParamKeyTable()),
		hooks:              nil,
		validatorCache:     make(map[string]cachedValidator, aminoCacheSize),
		validatorCacheList: list.New(),
		codespace:          codespace,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger { return ctx.Logger().With("module", "x/staking") }

// Set the validator hooks
func (k *Keeper) SetHooks(sh sdk.StakingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}
	k.hooks = sh
	return k
}

// return the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

// Load the last total validator power.
func (k Keeper) GetLastTotalPower(ctx sdk.Context) (power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(LastTotalPowerKey)
	if b == nil {
		return sdk.ZeroInt()
	}
	k.cdc.MustUnmarshalBinaryLengthPrefixed(b, &power)
	return
}

// Set the last total validator power.
func (k Keeper) SetLastTotalPower(ctx sdk.Context, power sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshalBinaryLengthPrefixed(power)
	store.Set(LastTotalPowerKey, b)
}

// freeCoinsToNotBonded adds coins to the not bonded pool of coins within the staking module
func (k Keeper) freeCoinsToNotBonded(ctx sdk.Context, amt sdk.Coins) sdk.Error {
	_, notBondedPool := k.GetPools(ctx)
	err := notBondedPool.SetCoins(notBondedPool.GetCoins().Add(amt))
	if err != nil {
		return sdk.ErrInternal(err.Error())
	}

	k.SetNotBondedPool(ctx, notBondedPool)
	return nil
}

// notBondedTokensToBonded transfers coins from the not bonded to the bonded pool within staking
func (k Keeper) notBondedTokensToBonded(ctx sdk.Context, notBondedTokens sdk.Int) {
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), notBondedTokens))
	err := k.supplyKeeper.SendCoinsPoolToPool(ctx, NotBondedTokensName, BondedTokensName, notBondedCoins)
	if err != nil {
		panic(err)
	}
}

// bondedTokensToNotBonded transfers coins from the bonded to the not bonded pool within staking
func (k Keeper) bondedTokensToNotBonded(ctx sdk.Context, bondedTokens sdk.Int) {
	bondedCoins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), bondedTokens))
	err := k.supplyKeeper.SendCoinsPoolToPool(ctx, BondedTokensName, NotBondedTokensName, bondedCoins)
	if err != nil {
		panic(err)
	}
}
