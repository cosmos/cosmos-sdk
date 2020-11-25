package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// Keeper of the supply store
type Keeper struct {
	cdc       *codec.Codec
	storeKey  sdk.StoreKey
	ak        types.AccountKeeper
	bk        types.BankKeeper
	permAddrs map[string]types.PermissionsForAddress
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ak types.AccountKeeper, bk types.BankKeeper, maccPerms map[string][]string) Keeper {
	// set the addresses
	permAddrs := make(map[string]types.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = types.NewPermissionsForAddress(name, perms)
	}

	return Keeper{
		cdc:       cdc,
		storeKey:  key,
		ak:        ak,
		bk:        bk,
		permAddrs: permAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetSupply retrieves the Supply from store
func (k Keeper) GetSupply(ctx sdk.Context) (supply exported.SupplyI) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, SupplyKey)
	defer iterator.Close()

	var totalSupply types.Supply
	for ; iterator.Valid(); iterator.Next() {
		var amount sdk.Dec
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &amount)
		totalSupply.Total = append(totalSupply.Total, sdk.NewCoin(string(iterator.Key()[1:]), amount))
	}

	return totalSupply
}

// SetSupply sets the Supply to store
func (k Keeper) SetSupply(ctx sdk.Context, supply exported.SupplyI) {
	tokensSupply := supply.GetTotal()
	for i := 0; i < len(tokensSupply); i++ {
		k.setTokenSupplyAmount(ctx, tokensSupply[i].Denom, tokensSupply[i].Amount)
	}
}

// ValidatePermissions validates that the module account has been granted
// permissions within its set of allowed permissions.
func (k Keeper) ValidatePermissions(macc exported.ModuleAccountI) error {
	permAddr := k.permAddrs[macc.GetName()]
	for _, perm := range macc.GetPermissions() {
		if !permAddr.HasPermission(perm) {
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}
	return nil
}

// setTokenSupplyAmount sets the supply amount of a token to the store
func (k Keeper) setTokenSupplyAmount(ctx sdk.Context, tokenSymbol string, supplyAmount sdk.Dec) {
	ctx.KVStore(k.storeKey).Set(getTokenSupplyKey(tokenSymbol), k.cdc.MustMarshalBinaryLengthPrefixed(supplyAmount))
}
